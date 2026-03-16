package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// =============================================================================
// Config helpers
// =============================================================================

// testAccRoutingDomainsDataSourceConfig returns a config that reads all
// routing domains without any pagination arguments.
func testAccRoutingDomainsDataSourceConfig() string {
	return `
data "spa_routing_domains" "all" {}
`
}

// testAccRoutingDomainsDataSourceBasicConfig creates one routing domain
// resource and reads all routing domains so the list is guaranteed non-empty.
func testAccRoutingDomainsDataSourceBasicConfig(fqdn string) string {
	return fmt.Sprintf(`
resource "spa_routing_domain" "basic_test" {
  fqdn         = %q
  type         = "internal"
  app_type     = "web"
  comment      = "Terraform acceptance test basic data source"
  flag         = "enabled"
  ip           = false
  location_ids = []
}

data "spa_routing_domains" "all" {
  depends_on = [spa_routing_domain.basic_test]
}
`, fqdn)
}

// testAccRoutingDomainsDataSourcePaginationConfig creates one routing domain
// resource and reads it back with offset=0, limit=1 so the pagination limit
// is verified against a known-present item.
func testAccRoutingDomainsDataSourcePaginationConfig(fqdn string) string {
	return fmt.Sprintf(`
resource "spa_routing_domain" "pagination_test" {
  fqdn         = %q
  type         = "internal"
  app_type     = "web"
  comment      = "Terraform acceptance test pagination"
  flag         = "enabled"
  ip           = false
  location_ids = []
}

data "spa_routing_domains" "paged" {
  offset     = 0
  limit      = 1
  depends_on = [spa_routing_domain.pagination_test]
}
`, fqdn)
}

// testAccRoutingDomainsDataSourceOffsetConfig creates two routing domain
// resources and exposes two overlapping paginated data source reads so the
// sliding-window relationship between offset=0/limit=2 and offset=1/limit=2
// can be verified.
func testAccRoutingDomainsDataSourceOffsetConfig(fqdn1, fqdn2 string) string {
	return fmt.Sprintf(`
resource "spa_routing_domain" "offset_1" {
  fqdn         = %q
  type         = "internal"
  app_type     = "web"
  comment      = "Terraform acceptance test offset domain 1"
  flag         = "enabled"
  ip           = false
  location_ids = []
}

resource "spa_routing_domain" "offset_2" {
  fqdn         = %q
  type         = "internal"
  app_type     = "web"
  comment      = "Terraform acceptance test offset domain 2"
  flag         = "enabled"
  ip           = false
  location_ids = []
}

data "spa_routing_domains" "page1" {
  offset     = 0
  limit      = 2
  depends_on = [spa_routing_domain.offset_1, spa_routing_domain.offset_2]
}

data "spa_routing_domains" "page2" {
  offset     = 1
  limit      = 2
  depends_on = [spa_routing_domain.offset_1, spa_routing_domain.offset_2]
}
`, fqdn1, fqdn2)
}

// testAccRoutingDomainsDataSourceRequiredFieldsConfig creates one routing
// domain resource and reads up to 5 items so that required fields on the
// first item can be verified against a known-present entry.
func testAccRoutingDomainsDataSourceRequiredFieldsConfig(fqdn string) string {
	return fmt.Sprintf(`
resource "spa_routing_domain" "fields_test" {
  fqdn         = %q
  type         = "internal"
  app_type     = "web"
  comment      = "Terraform acceptance test required fields"
  flag         = "enabled"
  ip           = false
  location_ids = []
}

data "spa_routing_domains" "paged" {
  offset     = 0
  limit      = 5
  depends_on = [spa_routing_domain.fields_test]
}
`, fqdn)
}

// testAccRoutingDomainsDataSourceConfigWithPagination returns a config that
// reads routing domains using explicit offset and limit values.
func testAccRoutingDomainsDataSourceConfigWithPagination(offset, limit int) string {
	return fmt.Sprintf(`
data "spa_routing_domains" "paged" {
  offset = %d
  limit  = %d
}
`, offset, limit)
}

// testAccRoutingDomainsDataSourceWithResourceConfig creates a routing domain
// resource and then reads all routing domains so we can verify the created
// domain appears in the list.
func testAccRoutingDomainsDataSourceWithResourceConfig(fqdn string) string {
	return fmt.Sprintf(`
resource "spa_routing_domain" "ds_test" {
  fqdn         = %q
  type         = "internal"
  app_type     = "web"
  comment      = "Terraform acceptance test data source"
  flag         = "enabled"
  ip           = false
  location_ids = []
}

data "spa_routing_domains" "all" {
  depends_on = [spa_routing_domain.ds_test]
}
`, fqdn)
}

// =============================================================================
// Custom check helpers
// =============================================================================

// testAccCheckRoutingDomainsContainsFQDN verifies that the given FQDN appears
// in the routing_domains list returned by a spa_routing_domains data source.
func testAccCheckRoutingDomainsContainsFQDN(dataSourceName, fqdn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found in state: %s", dataSourceName)
		}

		// Retrieve the total count of items in the list attribute.
		countStr, ok := rs.Primary.Attributes["routing_domains.#"]
		if !ok {
			return fmt.Errorf("routing_domains.# not found in state for %s", dataSourceName)
		}

		var count int
		if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
			return fmt.Errorf("could not parse routing_domains.# value %q: %w", countStr, err)
		}

		for i := 0; i < count; i++ {
			key := fmt.Sprintf("routing_domains.%d.fqdn", i)
			if rs.Primary.Attributes[key] == fqdn {
				return nil
			}
		}

		return fmt.Errorf("routing domain with FQDN %q not found in %s (checked %d items)", fqdn, dataSourceName, count)
	}
}

// testAccCheckRoutingDomainsCountPositive verifies that total is greater than
// zero for the given data source.
func testAccCheckRoutingDomainsCountPositive(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found in state: %s", dataSourceName)
		}

		var total int
		if _, err := fmt.Sscanf(rs.Primary.Attributes["total"], "%d", &total); err != nil {
			return fmt.Errorf("could not parse total: %w", err)
		}

		if total <= 0 {
			return fmt.Errorf("expected total > 0, got %d", total)
		}
		return nil
	}
}

// =============================================================================
// Tests
// =============================================================================

// TestAccRoutingDomainsDataSource_basic creates one routing domain and reads
// all routing domains without pagination args, verifying defaults and a
// non-empty list.
func TestAccRoutingDomainsDataSource_basic(t *testing.T) {
	fqdn := "tf-acc-ds-rd-basic.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupRoutingDomain(fqdn)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoutingDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingDomainsDataSourceBasicConfig(fqdn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("spa_routing_domain.basic_test"),
					// total must come back as a computed integer
					resource.TestCheckResourceAttrSet("data.spa_routing_domains.all", "total"),
					// pagination defaults are reflected in state
					resource.TestCheckResourceAttr("data.spa_routing_domains.all", "offset", "0"),
					resource.TestCheckResourceAttr("data.spa_routing_domains.all", "limit", "-1"),
					// list attribute must be present and non-empty
					resource.TestCheckResourceAttrSet("data.spa_routing_domains.all", "routing_domains.#"),
					testAccCheckRoutingDomainsCountPositive("data.spa_routing_domains.all"),
				),
			},
		},
	})
}

// TestAccRoutingDomainsDataSource_withResource creates a routing domain and
// then verifies it appears in the routing_domains list.
func TestAccRoutingDomainsDataSource_withResource(t *testing.T) {
	fqdn := "tf-acc-ds-routing-domains.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupRoutingDomain(fqdn)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoutingDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingDomainsDataSourceWithResourceConfig(fqdn),
				Check: resource.ComposeAggregateTestCheckFunc(
					// The resource must exist in the API
					testAccCheckRoutingDomainExistsInAPI("spa_routing_domain.ds_test"),
					// The data source must include the newly created domain
					testAccCheckRoutingDomainsContainsFQDN("data.spa_routing_domains.all", fqdn),
					// Sanity-check the meta fields
					testAccCheckRoutingDomainsCountPositive("data.spa_routing_domains.all"),
					resource.TestCheckResourceAttrSet("data.spa_routing_domains.all", "total"),
				),
			},
		},
	})
}

// TestAccRoutingDomainsDataSource_pagination creates one routing domain,
// requests a single page of limit=1 and verifies that pagination meta fields
// are set correctly and exactly 1 item is returned.
func TestAccRoutingDomainsDataSource_pagination(t *testing.T) {
	fqdn := "tf-acc-ds-rd-pagination.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupRoutingDomain(fqdn)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoutingDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingDomainsDataSourcePaginationConfig(fqdn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("spa_routing_domain.pagination_test"),
					resource.TestCheckResourceAttr("data.spa_routing_domains.paged", "offset", "0"),
					resource.TestCheckResourceAttr("data.spa_routing_domains.paged", "limit", "1"),
					// Exactly 1 item should be returned with limit=1
					resource.TestCheckResourceAttr("data.spa_routing_domains.paged", "routing_domains.#", "1"),
					resource.TestCheckResourceAttrSet("data.spa_routing_domains.paged", "total"),
				),
			},
		},
	})
}

// TestAccRoutingDomainsDataSource_offset creates two routing domains, then
// fetches two overlapping pages and confirms the sliding-window relationship:
// page1[offset=0,limit=2][1] must equal page2[offset=1,limit=2][0].
// Creating the domains explicitly makes the test self-contained and prevents
// a false positive when the environment has fewer than 2 existing domains.
func TestAccRoutingDomainsDataSource_offset(t *testing.T) {
	fqdn1 := "tf-acc-ds-rd-offset-1.example.com"
	fqdn2 := "tf-acc-ds-rd-offset-2.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupRoutingDomain(fqdn1)
			testAccCleanupRoutingDomain(fqdn2)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoutingDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingDomainsDataSourceOffsetConfig(fqdn1, fqdn2),
				Check: resource.ComposeAggregateTestCheckFunc(
					// both resources must exist
					testAccCheckRoutingDomainExistsInAPI("spa_routing_domain.offset_1"),
					testAccCheckRoutingDomainExistsInAPI("spa_routing_domain.offset_2"),
					// pagination attributes reflected in state
					resource.TestCheckResourceAttr("data.spa_routing_domains.page1", "offset", "0"),
					resource.TestCheckResourceAttr("data.spa_routing_domains.page1", "limit", "2"),
					resource.TestCheckResourceAttr("data.spa_routing_domains.page2", "offset", "1"),
					resource.TestCheckResourceAttr("data.spa_routing_domains.page2", "limit", "2"),
					resource.TestCheckResourceAttrSet("data.spa_routing_domains.page1", "total"),
					resource.TestCheckResourceAttrSet("data.spa_routing_domains.page2", "total"),
					// page1 must return exactly 2 items (limit=2 and >=2 domains exist)
					resource.TestCheckResourceAttr("data.spa_routing_domains.page1", "routing_domains.#", "2"),
					// sliding-window: page1[1] and page2[0] must be the same domain
					resource.TestCheckResourceAttrPair(
						"data.spa_routing_domains.page1", "routing_domains.1.fqdn",
						"data.spa_routing_domains.page2", "routing_domains.0.fqdn",
					),
				),
			},
		},
	})
}

// TestAccRoutingDomainsDataSource_eachItemHasFQDN creates one routing domain
// and verifies that the first item in the returned list has all required
// fields populated.
func TestAccRoutingDomainsDataSource_eachItemHasFQDN(t *testing.T) {
	fqdn := "tf-acc-ds-rd-fields.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupRoutingDomain(fqdn)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoutingDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingDomainsDataSourceRequiredFieldsConfig(fqdn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("spa_routing_domain.fields_test"),
					resource.TestCheckResourceAttrSet("data.spa_routing_domains.paged", "routing_domains.#"),
					// Verify first item has required fields populated
					resource.TestCheckResourceAttrSet("data.spa_routing_domains.paged", "routing_domains.0.fqdn"),
					resource.TestCheckResourceAttrSet("data.spa_routing_domains.paged", "routing_domains.0.type"),
				),
			},
		},
	})
}

// TestAccRoutingDomainsDataSource_apiDirectVerification creates a routing
// domain via the resource, then independently verifies via the API client that
// the domain is in the response returned by GetRoutingDomains.
func TestAccRoutingDomainsDataSource_apiDirectVerification(t *testing.T) {
	fqdn := "tf-acc-ds-api-verify.example.com"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupRoutingDomain(fqdn)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoutingDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingDomainsDataSourceWithResourceConfig(fqdn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingDomainExistsInAPI("spa_routing_domain.ds_test"),
					// Verify via API client directly that GetRoutingDomains returns our domain
					func(s *terraform.State) error {
						client, err := testAccCreateClient()
						if err != nil {
							return fmt.Errorf("failed to create API client: %w", err)
						}
						ctx := context.Background()
						result, err := client.GetRoutingDomains(ctx, 0, -1)
						if err != nil {
							return fmt.Errorf("GetRoutingDomains API call failed: %w", err)
						}
						if result.Total <= 0 {
							return fmt.Errorf("expected Total > 0, got %d", result.Total)
						}
						for _, rd := range result.RoutingDomains {
							if rd.FQDN == fqdn {
								return nil
							}
						}
						return fmt.Errorf("routing domain %q not found in GetRoutingDomains response (%d items)", fqdn, result.Total)
					},
				),
			},
		},
	})
}
