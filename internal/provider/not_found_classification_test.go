package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// "not found" must be decided by the parsed HTTP status code, never by
// searching the error text for "404".
//
// The error message embeds a client-generated v4 transaction ID (roughly 1 in
// 113 of which contain the substring "404") and the raw response body (which
// the SPA API fills with a verbatim echo of the caller's own input). The
// fixtures below reproduce both of those shapes with synthetic values.

// An HTTP 400 whose transaction ID happens to end in "404a".
func poisonedTxIDError() *APIError {
	return &APIError{
		StatusCode:    http.StatusBadRequest,
		TransactionID: "3f2a1c7e-9b4d-4e18-a6c5-1d2e3f404a7b",
		Body:          `{"type": "https://errors-api.cloud.com/common/Error", "detail": "Input payload validation failed"}`,
	}
}

// A 500 whose body mentions "404" for reasons unrelated to the status.
func poisonedBodyError() *APIError {
	return &APIError{
		StatusCode:    http.StatusInternalServerError,
		TransactionID: "11111111-2222-3333-4444-555555555555",
		Body:          `{"message":"upstream routing service unavailable","detail":"gateway pool node-404 timed out"}`,
	}
}

func TestAPIErrorMessageFormatIsUnchanged(t *testing.T) {
	err := poisonedTxIDError()
	want := fmt.Sprintf("API request failed with status %d (transaction ID: %s): %s",
		err.StatusCode, err.TransactionID, err.Body)
	if got := err.Error(); got != want {
		t.Errorf("APIError.Error() changed the wire-format message\n got: %s\nwant: %s", got, want)
	}
}

func TestIsNotFoundUsesStatusCodeNotErrorText(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
		// poisoned marks a case whose message text contains "404" while the
		// response is not a 404 — the exact shape the old substring check got
		// wrong. Asserted below, so a fixture that quietly stops containing
		// "404" fails the test instead of silently hollowing it out.
		poisoned bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "genuine 404", err: &APIError{StatusCode: http.StatusNotFound, TransactionID: "abc", Body: `{"detail":"Not Found"}`}, want: true},
		{name: "404 whose text never mentions the number", err: &APIError{StatusCode: http.StatusNotFound, TransactionID: "abc", Body: "gone"}, want: true},
		{name: "400 with 404 in the transaction ID", err: poisonedTxIDError(), want: false, poisoned: true},
		{name: "500 with 404 in the body", err: poisonedBodyError(), want: false, poisoned: true},
		{name: "401 echoing a transaction ID containing 404", err: &APIError{
			StatusCode:    http.StatusUnauthorized,
			TransactionID: "7c1b5e90-2f7d-404a-9c9e-8a3d6b2f17c4",
			Body:          `{"detail": "Validate Identity Error: <Invalid Bearer Token>", "parameters": [{"name": "Citrix-TransactionId", "value": "7c1b5e90-2f7d-404a-9c9e-8a3d6b2f17c4"}]}`,
		}, want: false, poisoned: true},
		{name: "400 echoing a caller value containing 404", err: &APIError{
			StatusCode:    http.StatusBadRequest,
			TransactionID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			Body:          `{"detail": "Input payload validation failed", "parameters": [{"name": "url", "value": "'error404.acme.com' is not valid"}]}`,
		}, want: false, poisoned: true},
		{name: "transport failure", err: errors.New("failed to make request (transaction ID: 404): connection reset"), want: false, poisoned: true},
		{name: "wrapped 404", err: fmt.Errorf("reading routing domain: %w", &APIError{StatusCode: http.StatusNotFound}), want: true},
	}

	poisonedSeen := 0
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsNotFound(tc.err); got != tc.want {
				t.Errorf("IsNotFound() = %v, want %v (error text: %v)", got, tc.want, tc.err)
			}
			if !tc.poisoned {
				return
			}
			poisonedSeen++
			// Guard the regression itself. These cases only prove anything
			// while their text really does contain "404" and the response
			// really is not a 404 — otherwise the old substring check would
			// have passed them too.
			if tc.want {
				t.Fatalf("a poisoned case cannot also be a genuine not-found")
			}
			if !strings.Contains(tc.err.Error(), "404") {
				t.Errorf("fixture no longer contains %q, so it no longer exercises the old substring check: %v", "404", tc.err)
			}
		})
	}

	if poisonedSeen == 0 {
		t.Error("no poisoned fixtures left: this test no longer guards the substring regression")
	}
}

func TestHandleResponseReturnsAPIErrorCarryingTheStatus(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError} {
		client := newTestClientWithStatus(status)
		_, err := client.GetRoutingDomain(context.Background(), "intranet.example.com")
		if status == http.StatusNotFound {
			if !IsNotFound(err) {
				t.Errorf("status %d: expected IsNotFound, got %v", status, err)
			}
		} else if IsNotFound(err) {
			t.Errorf("status %d: must not be classified as not-found", status)
		}

		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("status %d: expected an *APIError, got %T (%v)", status, err, err)
		}
		if apiErr.StatusCode != status {
			t.Errorf("expected StatusCode %d, got %d", status, apiErr.StatusCode)
		}
	}
}

// --- the acceptance criterion, per resource ---------------------------------

// notFoundMockClient fails every read with the same error, so each resource's
// Read is exercised against a single, deliberately poisoned API failure.
type notFoundMockClient struct {
	SPAClient
	err error
}

func (m *notFoundMockClient) GetRoutingDomain(context.Context, string) (*RoutingDomain, error) {
	return nil, m.err
}
func (m *notFoundMockClient) GetApplication(context.Context, string) (*Application, error) {
	return nil, m.err
}
func (m *notFoundMockClient) GetApplicationAwaitSSO(context.Context, string) (*Application, error) {
	return nil, m.err
}
func (m *notFoundMockClient) GetAccessPolicy(context.Context, string) (*AccessPolicy, error) {
	return nil, m.err
}
func (m *notFoundMockClient) GetSecurityGroup(context.Context, string) (*SecurityGroup, error) {
	return nil, m.err
}
func (m *notFoundMockClient) GetSessionPolicy(context.Context, string) (*SessionPolicy, error) {
	return nil, m.err
}

// stateWithOnlyKey builds a state for the given schema in which every attribute
// is null except the resource's identifying one — enough to reach the API call.
func stateWithOnlyKey(ctx context.Context, t *testing.T, s schema.Schema, keyAttr, keyValue string) tfsdk.State {
	t.Helper()
	objType, ok := s.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("schema type is not an object")
	}
	values := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	if _, ok := objType.AttributeTypes[keyAttr]; !ok {
		t.Fatalf("schema has no attribute %q", keyAttr)
	}
	values[keyAttr] = tftypes.NewValue(tftypes.String, keyValue)
	return tfsdk.State{Schema: s, Raw: tftypes.NewValue(objType, values)}
}

// notFoundClassificationResource is one resource whose Read must classify
// failures by status code. Both tests below iterate this single table, so a
// resource can never be covered by one of them and forgotten by the other.
type notFoundClassificationResource struct {
	name     string
	newRes   func(client SPAClient) resource.Resource
	keyAttr  string
	keyValue string
}

var notFoundClassificationResources = []notFoundClassificationResource{
	{"citrixspa_routing_domain", func(c SPAClient) resource.Resource { return &RoutingDomainResource{client: c} }, "fqdn", "intranet.example.com"},
	{"citrixspa_application", func(c SPAClient) resource.Resource { return &ApplicationResource{client: c} }, "id", "app-123"},
	{"citrixspa_access_policy", func(c SPAClient) resource.Resource { return &AccessPolicyResource{client: c} }, "id", "policy-123"},
	{"citrixspa_security_group", func(c SPAClient) resource.Resource { return &SecurityGroupResource{client: c} }, "id", "sg-123"},
	{"citrixspa_session_policy", func(c SPAClient) resource.Resource { return &SessionPolicyResource{client: c} }, "id", "sp-123"},
}

// TestNon404ErrorDoesNotRemoveResourceFromState is the regression guard: a
// failure that is not an HTTP 404 must surface as an error and leave the
// resource in state, however its text happens to read.
func TestNon404ErrorDoesNotRemoveResourceFromState(t *testing.T) {
	ctx := context.Background()

	resources := notFoundClassificationResources

	failures := []struct {
		name string
		err  error
	}{
		{"http 500 with 404 in the body", poisonedBodyError()},
		{"http 400 with 404 in the transaction ID", poisonedTxIDError()},
	}

	for _, rc := range resources {
		for _, f := range failures {
			t.Run(rc.name+"/"+f.name, func(t *testing.T) {
				res := rc.newRes(&notFoundMockClient{err: f.err})

				schemaResp := &resource.SchemaResponse{}
				res.Schema(ctx, resource.SchemaRequest{}, schemaResp)
				if schemaResp.Diagnostics.HasError() {
					t.Fatalf("schema: %v", schemaResp.Diagnostics.Errors())
				}

				state := stateWithOnlyKey(ctx, t, schemaResp.Schema, rc.keyAttr, rc.keyValue)
				readResp := &resource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema, Raw: state.Raw}}
				res.Read(ctx, resource.ReadRequest{State: state}, readResp)

				if readResp.State.Raw.IsNull() {
					t.Fatalf("resource was removed from state on a non-404 failure: %v", f.err)
				}
				if !readResp.Diagnostics.HasError() {
					t.Errorf("expected an error diagnostic, got: %v", readResp.Diagnostics)
				}
				for _, d := range readResp.Diagnostics.Warnings() {
					if strings.Contains(d.Summary(), "Deleted Outside Terraform") {
						t.Errorf("non-404 failure wrongly reported as an out-of-band deletion: %s", d.Summary())
					}
				}
			})
		}
	}
}

// TestGenuine404StillRemovesResourceFromState pins the other half of the
// acceptance criteria: real 404s keep behaving as drift.
func TestGenuine404StillRemovesResourceFromState(t *testing.T) {
	ctx := context.Background()
	notFound := &APIError{StatusCode: http.StatusNotFound, TransactionID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", Body: `{"detail":"Not Found"}`}

	for _, rc := range notFoundClassificationResources {
		t.Run(rc.name, func(t *testing.T) {
			res := rc.newRes(&notFoundMockClient{err: notFound})

			schemaResp := &resource.SchemaResponse{}
			res.Schema(ctx, resource.SchemaRequest{}, schemaResp)
			if schemaResp.Diagnostics.HasError() {
				t.Fatalf("schema: %v", schemaResp.Diagnostics.Errors())
			}

			state := stateWithOnlyKey(ctx, t, schemaResp.Schema, rc.keyAttr, rc.keyValue)
			readResp := &resource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema, Raw: state.Raw}}
			res.Read(ctx, resource.ReadRequest{State: state}, readResp)

			if !readResp.State.Raw.IsNull() {
				t.Errorf("a genuine 404 must still remove the resource from state")
			}
			if readResp.Diagnostics.HasError() {
				t.Errorf("a genuine 404 must not raise an error diagnostic: %v", readResp.Diagnostics.Errors())
			}
		})
	}
}

// --- orphan recovery must key off the status too ----------------------------

// orphanRecoveryMockClient fails every create with the same error and records
// whether the orphan search ran. A non-nil existing application stands in for
// one that was already on the tenant before this apply.
type orphanRecoveryMockClient struct {
	SPAClient
	createErr error
	existing  *ApplicationListItem

	searched bool
}

func (m *orphanRecoveryMockClient) CreateApplication(context.Context, *Application) (*Application, error) {
	return nil, m.createErr
}

func (m *orphanRecoveryMockClient) GetApplications(_ context.Context, _, _ int, name, appType string) (*ApplicationsResponse, error) {
	m.searched = true
	if m.existing != nil && m.existing.Name == name && m.existing.Type == appType {
		return &ApplicationsResponse{Applications: []ApplicationListItem{*m.existing}}, nil
	}
	return &ApplicationsResponse{}, nil
}

// runApplicationCreate drives the real ApplicationResource.Create against the
// given client and returns the response, so the assertions below are about
// production behaviour rather than a reimplementation of it.
func runApplicationCreate(t *testing.T, client SPAClient) *resource.CreateResponse {
	t.Helper()
	ctx := context.Background()
	r := &ApplicationResource{client: client}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema: %v", schemaResp.Diagnostics.Errors())
	}

	model := testApplicationState()
	model.ID = types.StringNull()

	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	if d := plan.Set(ctx, &model); d.HasError() {
		t.Fatalf("plan set: %v", d)
	}

	createResp := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, createResp)
	return createResp
}

// The SPA API echoes the caller's own payload back in validation errors, so a
// bad URL containing the literal text "status 500" produces a 400 whose message
// reads like a 500. Same shape as the poisoned "404" fixtures above.
func poisonedStatus500Error() *APIError {
	return &APIError{
		StatusCode:    http.StatusBadRequest,
		TransactionID: "5a1c9e20-7b3d-4f61-9a2c-6e8b1d4f3a27",
		Body:          `{"detail": "Input payload validation failed", "parameters": [{"name": "url", "value": "'https://portal.example.com/status 500' is not a valid URL"}]}`,
	}
}

func TestOrphanRecoveryUsesStatusCodeNotErrorText(t *testing.T) {
	poisoned := poisonedStatus500Error()
	if !strings.Contains(poisoned.Error(), "status 500") {
		t.Fatalf("fixture no longer contains %q, so it no longer exercises the old substring check", "status 500")
	}

	m := &orphanRecoveryMockClient{createErr: poisoned}
	createResp := runApplicationCreate(t, m)

	if !createResp.Diagnostics.HasError() {
		t.Error("a failed create must surface an error diagnostic")
	}
	if m.searched {
		t.Errorf("orphan recovery ran for an HTTP %d whose text merely contains %q",
			poisoned.StatusCode, "status 500")
	}
}

// The damage the above prevents: without a status check, a validation error
// starts an orphan hunt that can adopt an unrelated application which merely
// shares the planned name and type, binding someone else's resource into this
// configuration's state.
func TestOrphanRecoveryDoesNotAdoptAForeignApplicationOnNon500(t *testing.T) {
	planned := testApplicationState()
	m := &orphanRecoveryMockClient{
		createErr: poisonedStatus500Error(),
		existing: &ApplicationListItem{
			ID:   "pre-existing-application",
			Name: planned.Name.ValueString(),
			Type: planned.Type.ValueString(),
		},
	}

	createResp := runApplicationCreate(t, m)

	if createResp.State.Raw.IsNull() {
		return // nothing was written to state at all, which is the desired outcome
	}
	var adopted ApplicationResourceModel
	if d := createResp.State.Get(context.Background(), &adopted); d.HasError() {
		return // no decodable state written
	}
	if adopted.ID.ValueString() == "pre-existing-application" {
		t.Errorf("a failed create adopted an unrelated existing application into state as id=%q",
			adopted.ID.ValueString())
	}
}

// The other half: a genuine 500 may really have left a half-created
// application behind, so recovery must still run and still adopt it.
func TestGenuine500StillRecoversOrphanedApplication(t *testing.T) {
	planned := testApplicationState()
	m := &orphanRecoveryMockClient{
		createErr: &APIError{
			StatusCode:    http.StatusInternalServerError,
			TransactionID: "bbbbbbbb-cccc-dddd-eeee-ffffffffffff",
			Body:          `{"detail":"Internal Server Error"}`,
		},
		existing: &ApplicationListItem{
			ID:   "orphaned-application",
			Name: planned.Name.ValueString(),
			Type: planned.Type.ValueString(),
		},
	}

	createResp := runApplicationCreate(t, m)

	if !m.searched {
		t.Fatal("a genuine 500 must still trigger the orphan recovery search")
	}
	if createResp.State.Raw.IsNull() {
		t.Fatal("a recovered orphan must be saved to state")
	}
	var recovered ApplicationResourceModel
	if d := createResp.State.Get(context.Background(), &recovered); d.HasError() {
		t.Fatalf("state get: %v", d)
	}
	if recovered.ID.ValueString() != "orphaned-application" {
		t.Errorf("expected the orphan to be adopted, got id=%q", recovered.ID.ValueString())
	}
}
