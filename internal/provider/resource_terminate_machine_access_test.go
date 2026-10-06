package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// =============================================================================
// Terminate Machine Access Resource Tests
// =============================================================================
//
// Like the user variant, the terminate-machine API validates the objectId
// *format* ("OID:/citrix/<uuid>"), not directory existence, so a synthetic
// identity drives a full create/verify/destroy lifecycle without a real machine.

// testAccTerminateMachineAccessConfig generates HCL for a
// citrixspa_terminate_machine_access resource.
func testAccTerminateMachineAccessConfig(resourceName, accountName, name, objectID, idpType string, duration int) string {
	if resourceName == "" {
		resourceName = "test"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "resource \"citrixspa_terminate_machine_access\" %q {\n", resourceName)
	fmt.Fprintf(&b, "  account_name = %q\n", accountName)
	fmt.Fprintf(&b, "  name         = %q\n", name)
	fmt.Fprintf(&b, "  object_id    = %q\n", objectID)
	fmt.Fprintf(&b, "  idp_type     = %q\n", idpType)
	fmt.Fprintf(&b, "  duration     = %d\n", duration)
	fmt.Fprintf(&b, "}\n")
	return b.String()
}

// testAccCheckTerminateMachineAccessDestroy verifies that all
// citrixspa_terminate_machine_access resources have been removed from the API.
func testAccCheckTerminateMachineAccessDestroy(s *terraform.State) error {
	client, err := testAccCreateClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "citrixspa_terminate_machine_access" {
			continue
		}
		id := rs.Primary.Attributes["id"]
		// Query the list directly so a real API error (auth/network/list) fails
		// the check rather than being mistaken for a successful deletion.
		machines, err := client.GetTerminateMachineAccess(ctx, 0, -1)
		if err != nil {
			return fmt.Errorf("failed to list terminate machine access records: %w", err)
		}
		for _, m := range machines.Items {
			if m.ID == id {
				return fmt.Errorf("terminate machine access %s still exists in the API after destroy", id)
			}
		}
	}
	return nil
}

// testAccCheckTerminateMachineAccessExistsInAPI verifies the record exists in
// the API and its object_id matches the Terraform state.
func testAccCheckTerminateMachineAccessExistsInAPI(resourceName string) resource.TestCheckFunc {
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
		rec, err := client.GetTerminateMachineAccessByID(context.Background(), id)
		if err != nil {
			return fmt.Errorf("terminate machine access %s not found in API: %s", id, err)
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

// TestAccTerminateMachineAccessResource_basic exercises the create, verify, and
// destroy lifecycle using a synthetic object identifier.
func TestAccTerminateMachineAccessResource_basic(t *testing.T) {
	oid := testAccRandomOID(t)
	resAddr := "citrixspa_terminate_machine_access.test"
	config := testAccTerminateMachineAccessConfig(
		"test", "tf-acc-terminate-machine", "tf-acc-machine", oid, "AD", 1,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTerminateMachineAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTerminateMachineAccessExistsInAPI(resAddr),
					resource.TestCheckResourceAttrSet(resAddr, "id"),
					resource.TestCheckResourceAttr(resAddr, "object_id", oid),
				),
			},
		},
	})
}
