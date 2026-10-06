package provider

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// =============================================================================
// Terminate User Access Resource Tests
// =============================================================================
//
// The terminate-access API validates the objectId *format* (it must be
// "OID:/citrix/<uuid>"), not directory existence, so a synthetic identity
// drives a full create/verify/destroy lifecycle without a real user.

// testAccRandomOID returns a synthetic, well-formed Citrix object identifier
// ("OID:/citrix/<uuid>") suitable for terminate-access resources.
func testAccRandomOID(t *testing.T) string {
	t.Helper()
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("failed to generate random OID: %s", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	guid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	return "OID:/citrix/" + guid
}

// testAccTerminateUserAccessConfig generates HCL for a
// citrixspa_terminate_user_access resource.
func testAccTerminateUserAccessConfig(resourceName, accountName, email, domain, objectID, idpType string, duration int) string {
	if resourceName == "" {
		resourceName = "test"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "resource \"citrixspa_terminate_user_access\" %q {\n", resourceName)
	fmt.Fprintf(&b, "  account_name = %q\n", accountName)
	fmt.Fprintf(&b, "  email        = %q\n", email)
	fmt.Fprintf(&b, "  domain_name  = %q\n", domain)
	fmt.Fprintf(&b, "  object_id    = %q\n", objectID)
	fmt.Fprintf(&b, "  idp_type     = %q\n", idpType)
	fmt.Fprintf(&b, "  duration     = %d\n", duration)
	fmt.Fprintf(&b, "}\n")
	return b.String()
}

// testAccCheckTerminateUserAccessDestroy verifies that all
// citrixspa_terminate_user_access resources have been removed from the API.
func testAccCheckTerminateUserAccessDestroy(s *terraform.State) error {
	client, err := testAccCreateClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "citrixspa_terminate_user_access" {
			continue
		}
		id := rs.Primary.Attributes["id"]
		// Query the list directly so a real API error (auth/network/list) fails
		// the check rather than being mistaken for a successful deletion.
		users, err := client.GetTerminateUserAccess(ctx, 0, -1)
		if err != nil {
			return fmt.Errorf("failed to list terminate user access records: %w", err)
		}
		for _, u := range users.Items {
			if u.ID == id {
				return fmt.Errorf("terminate user access %s still exists in the API after destroy", id)
			}
		}
	}
	return nil
}

// testAccCheckTerminateUserAccessExistsInAPI verifies the record exists in the
// API and its object_id matches the Terraform state.
func testAccCheckTerminateUserAccessExistsInAPI(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}
		id := rs.Primary.Attributes["id"]
		if id == "" {
			return fmt.Errorf("no ID set for resource %s", resourceName)
		}

		client, err := testAccCreateClient()
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}
		rec, err := client.GetTerminateUserAccessByID(context.Background(), id)
		if err != nil {
			return fmt.Errorf("terminate user access %s not found in API: %s", id, err)
		}
		if rec.ObjectID != rs.Primary.Attributes["object_id"] {
			return fmt.Errorf("object_id mismatch: API=%q, state=%q", rec.ObjectID, rs.Primary.Attributes["object_id"])
		}
		return nil
	}
}

// =============================================================================
// Tests
// =============================================================================

// TestAccTerminateUserAccessResource_basic exercises the create, verify, and
// destroy lifecycle using a synthetic object identifier.
func TestAccTerminateUserAccessResource_basic(t *testing.T) {
	oid := testAccRandomOID(t)
	resAddr := "citrixspa_terminate_user_access.test"
	config := testAccTerminateUserAccessConfig(
		"test", "tf-acc-terminate-user", "tf-acc-terminate-user@example.invalid",
		"example.invalid", oid, "AD", 1,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTerminateUserAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTerminateUserAccessExistsInAPI(resAddr),
					resource.TestCheckResourceAttrSet(resAddr, "id"),
					resource.TestCheckResourceAttr(resAddr, "object_id", oid),
				),
			},
		},
	})
}
