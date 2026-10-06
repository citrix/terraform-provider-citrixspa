package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// resourceSchemaAttributes returns the top-level schema attributes for a resource.
func resourceSchemaAttributes(t *testing.T, r resource.Resource) map[string]schema.Attribute {
	t.Helper()
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema build failed: %v", resp.Diagnostics.Errors())
	}
	return resp.Schema.Attributes
}

// TestTerminateMachineAccessNameOptional verifies that the machine name is
// optional. The SPA service (and its OpenAPI spec) requires only account_name,
// object_id and idp_type to identify a machine; name is not required.
func TestTerminateMachineAccessNameOptional(t *testing.T) {
	attrs := resourceSchemaAttributes(t, NewTerminateMachineAccessResource())

	if got := attrs["name"]; got.IsRequired() {
		t.Errorf("expected terminate_machine_access.name to be optional, but it is required")
	}
	if got := attrs["name"]; !got.IsOptional() {
		t.Errorf("expected terminate_machine_access.name to be optional")
	}
	for _, name := range []string{"account_name", "object_id", "idp_type"} {
		if !attrs[name].IsRequired() {
			t.Errorf("expected terminate_machine_access.%s to remain required", name)
		}
	}
}

// TestTerminateUserAccessOptionalFields verifies that email, domain_name and
// duration are optional (the SPA service does not require them), while
// account_name, object_id and idp_type remain required.
func TestTerminateUserAccessOptionalFields(t *testing.T) {
	attrs := resourceSchemaAttributes(t, NewTerminateUserAccessResource())

	for _, name := range []string{"email", "domain_name", "duration"} {
		if attrs[name].IsRequired() {
			t.Errorf("expected terminate_user_access.%s to be optional, but it is required", name)
		}
		if !attrs[name].IsOptional() {
			t.Errorf("expected terminate_user_access.%s to be optional", name)
		}
	}

	for _, name := range []string{"account_name", "object_id", "idp_type"} {
		if !attrs[name].IsRequired() {
			t.Errorf("expected terminate_user_access.%s to remain required", name)
		}
	}
}

// mockTerminateClient embeds SPAClient so any unimplemented method panics if hit;
// the terminate Create paths only use the create + get-by-id methods, which are
// overridden per test.
type mockTerminateClient struct {
	SPAClient
	createMachine  func(ctx context.Context, m *TerminateMachineAccess) (*TerminateMachineAccess, error)
	getMachineByID func(ctx context.Context, id string) (*TerminateMachineAccess, error)
	createUser     func(ctx context.Context, u *TerminateUserAccess) (*TerminateUserAccess, error)
	getUserByID    func(ctx context.Context, id string) (*TerminateUserAccess, error)
}

func (m *mockTerminateClient) CreateTerminateMachineAccess(ctx context.Context, machine *TerminateMachineAccess) (*TerminateMachineAccess, error) {
	return m.createMachine(ctx, machine)
}

func (m *mockTerminateClient) GetTerminateMachineAccessByID(ctx context.Context, id string) (*TerminateMachineAccess, error) {
	return m.getMachineByID(ctx, id)
}

func (m *mockTerminateClient) CreateTerminateUserAccess(ctx context.Context, user *TerminateUserAccess) (*TerminateUserAccess, error) {
	return m.createUser(ctx, user)
}

func (m *mockTerminateClient) GetTerminateUserAccessByID(ctx context.Context, id string) (*TerminateUserAccess, error) {
	return m.getUserByID(ctx, id)
}

// runMachineCreate builds a plan from the model, runs Create against the mock,
// and returns the resulting state model plus diagnostics.
func runMachineCreate(t *testing.T, client SPAClient, plan TerminateMachineAccessResourceModel) (TerminateMachineAccessResourceModel, diag.Diagnostics) {
	t.Helper()
	ctx := context.Background()
	r := &TerminateMachineAccessResource{client: client}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	planValue := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := planValue.Set(ctx, &plan); diags.HasError() {
		t.Fatalf("failed to set plan: %v", diags.Errors())
	}

	createResp := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: planValue}, createResp)

	var out TerminateMachineAccessResourceModel
	createResp.State.Get(ctx, &out)
	return out, createResp.Diagnostics
}

// runUserCreate is the terminate_user_access equivalent of runMachineCreate.
func runUserCreate(t *testing.T, client SPAClient, plan TerminateUserAccessResourceModel) (TerminateUserAccessResourceModel, diag.Diagnostics) {
	t.Helper()
	ctx := context.Background()
	r := &TerminateUserAccessResource{client: client}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	planValue := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := planValue.Set(ctx, &plan); diags.HasError() {
		t.Fatalf("failed to set plan: %v", diags.Errors())
	}

	createResp := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: planValue}, createResp)

	var out TerminateUserAccessResourceModel
	createResp.State.Get(ctx, &out)
	return out, createResp.Diagnostics
}

// TestTerminateMachineAccessCreatePreservesPlanValues verifies that Create keeps
// user-configured (known) values and only fills omitted (unknown) optional
// fields from the post-create re-read.
func TestTerminateMachineAccessCreatePreservesPlanValues(t *testing.T) {
	created := &TerminateMachineAccess{ID: "machine-123", ObjectID: "OID:/ad/host1"}
	stored := &TerminateMachineAccess{
		ID: "machine-123", AccountName: "SERVER-ACCOUNT", Name: "server-name",
		DNSHostName: "server.dns", DomainName: "server.domain",
		ObjectID: "OID:/ad/host1", IDPType: "AD", Duration: 99,
	}
	client := &mockTerminateClient{
		createMachine: func(ctx context.Context, m *TerminateMachineAccess) (*TerminateMachineAccess, error) {
			return created, nil
		},
		getMachineByID: func(ctx context.Context, id string) (*TerminateMachineAccess, error) { return stored, nil },
	}

	plan := TerminateMachineAccessResourceModel{
		ID:          types.StringUnknown(),
		AccountName: types.StringValue("PLAN-ACCOUNT"),
		Name:        types.StringValue("plan-name"),
		DNSHostName: types.StringUnknown(),
		DomainName:  types.StringUnknown(),
		ObjectID:    types.StringValue("OID:/ad/host1"),
		IDPType:     types.StringValue("AD"),
		Duration:    types.Int64Value(30),
	}

	out, diags := runMachineCreate(t, client, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if out.ID.ValueString() != "machine-123" {
		t.Errorf("id: got %q, want machine-123", out.ID.ValueString())
	}
	// Required + explicitly-set optional plan values are preserved.
	if out.AccountName.ValueString() != "PLAN-ACCOUNT" {
		t.Errorf("account_name: got %q, want PLAN-ACCOUNT (preserved)", out.AccountName.ValueString())
	}
	if out.Name.ValueString() != "plan-name" {
		t.Errorf("name: got %q, want plan-name (should not be overwritten by re-read)", out.Name.ValueString())
	}
	if out.Duration.ValueInt64() != 30 {
		t.Errorf("duration: got %d, want 30 (preserved)", out.Duration.ValueInt64())
	}
	// Omitted (unknown) optional values are resolved from the stored record.
	if out.DNSHostName.ValueString() != "server.dns" {
		t.Errorf("dns_host_name: got %q, want server.dns (from re-read)", out.DNSHostName.ValueString())
	}
	if out.DomainName.ValueString() != "server.domain" {
		t.Errorf("domain_name: got %q, want server.domain (from re-read)", out.DomainName.ValueString())
	}
}

// TestTerminateMachineAccessCreateFallsBackWhenReReadMisses verifies that a
// re-read miss (record not yet visible in the list API) is non-fatal and the
// provider falls back to the create response.
func TestTerminateMachineAccessCreateFallsBackWhenReReadMisses(t *testing.T) {
	created := &TerminateMachineAccess{ID: "machine-xyz", Name: "created-name", ObjectID: "OID:/ad/host2", Duration: 7}
	client := &mockTerminateClient{
		createMachine: func(ctx context.Context, m *TerminateMachineAccess) (*TerminateMachineAccess, error) {
			return created, nil
		},
		getMachineByID: func(ctx context.Context, id string) (*TerminateMachineAccess, error) {
			return nil, fmt.Errorf("terminate machine access with ID %s not found", id)
		},
	}

	plan := TerminateMachineAccessResourceModel{
		ID:          types.StringUnknown(),
		AccountName: types.StringValue("ACC"),
		Name:        types.StringUnknown(),
		DNSHostName: types.StringUnknown(),
		DomainName:  types.StringUnknown(),
		ObjectID:    types.StringValue("OID:/ad/host2"),
		IDPType:     types.StringValue("AD"),
		Duration:    types.Int64Unknown(),
	}

	out, diags := runMachineCreate(t, client, plan)
	if diags.HasError() {
		t.Fatalf("expected no error on re-read miss (fallback), got: %v", diags.Errors())
	}
	if out.ID.ValueString() != "machine-xyz" {
		t.Errorf("id: got %q, want machine-xyz (from create response)", out.ID.ValueString())
	}
	if out.Name.ValueString() != "created-name" {
		t.Errorf("name: got %q, want created-name (fallback)", out.Name.ValueString())
	}
	if out.Duration.ValueInt64() != 7 {
		t.Errorf("duration: got %d, want 7 (fallback)", out.Duration.ValueInt64())
	}
}

// TestTerminateUserAccessCreatePreservesPlanValues mirrors the machine test for
// the user resource.
func TestTerminateUserAccessCreatePreservesPlanValues(t *testing.T) {
	created := &TerminateUserAccess{ID: "user-1", ObjectID: "OID:/ad/u1"}
	stored := &TerminateUserAccess{
		ID: "user-1", AccountName: "SERVER-ACC", Email: "server@x.com",
		DomainName: "server.domain", ObjectID: "OID:/ad/u1", IDPType: "AD", Duration: 88,
	}
	client := &mockTerminateClient{
		createUser:  func(ctx context.Context, u *TerminateUserAccess) (*TerminateUserAccess, error) { return created, nil },
		getUserByID: func(ctx context.Context, id string) (*TerminateUserAccess, error) { return stored, nil },
	}

	plan := TerminateUserAccessResourceModel{
		ID:          types.StringUnknown(),
		AccountName: types.StringValue("PLAN-ACC"),
		Email:       types.StringValue("plan@x.com"),
		DomainName:  types.StringUnknown(),
		ObjectID:    types.StringValue("OID:/ad/u1"),
		IDPType:     types.StringValue("AD"),
		Duration:    types.Int64Value(15),
	}

	out, diags := runUserCreate(t, client, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if out.ID.ValueString() != "user-1" {
		t.Errorf("id: got %q, want user-1", out.ID.ValueString())
	}
	if out.Email.ValueString() != "plan@x.com" {
		t.Errorf("email: got %q, want plan@x.com (preserved)", out.Email.ValueString())
	}
	if out.Duration.ValueInt64() != 15 {
		t.Errorf("duration: got %d, want 15 (preserved)", out.Duration.ValueInt64())
	}
	if out.DomainName.ValueString() != "server.domain" {
		t.Errorf("domain_name: got %q, want server.domain (from re-read)", out.DomainName.ValueString())
	}
}

// TestTerminateUserAccessCreateFallsBackWhenReReadMisses mirrors the machine
// fallback test for the user resource.
func TestTerminateUserAccessCreateFallsBackWhenReReadMisses(t *testing.T) {
	created := &TerminateUserAccess{ID: "user-9", Email: "created@x.com", ObjectID: "OID:/ad/u9", Duration: 5}
	client := &mockTerminateClient{
		createUser: func(ctx context.Context, u *TerminateUserAccess) (*TerminateUserAccess, error) { return created, nil },
		getUserByID: func(ctx context.Context, id string) (*TerminateUserAccess, error) {
			return nil, fmt.Errorf("terminate user access with ID %s not found", id)
		},
	}

	plan := TerminateUserAccessResourceModel{
		ID:          types.StringUnknown(),
		AccountName: types.StringValue("ACC"),
		Email:       types.StringUnknown(),
		DomainName:  types.StringUnknown(),
		ObjectID:    types.StringValue("OID:/ad/u9"),
		IDPType:     types.StringValue("AD"),
		Duration:    types.Int64Unknown(),
	}

	out, diags := runUserCreate(t, client, plan)
	if diags.HasError() {
		t.Fatalf("expected no error on re-read miss (fallback), got: %v", diags.Errors())
	}
	if out.ID.ValueString() != "user-9" {
		t.Errorf("id: got %q, want user-9 (from create response)", out.ID.ValueString())
	}
	if out.Email.ValueString() != "created@x.com" {
		t.Errorf("email: got %q, want created@x.com (fallback)", out.Email.ValueString())
	}
	if out.Duration.ValueInt64() != 5 {
		t.Errorf("duration: got %d, want 5 (fallback)", out.Duration.ValueInt64())
	}
}
