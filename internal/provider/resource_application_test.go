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
// Application Tests
// =============================================================================

const testAppIcon = "iVBORw0KGgoAAAANSUhEUgAAAAgAAAAICAYAAADED76LAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsIAAA7CARUoSoAAAAAaSURBVChTY6AzeCuj8h+EoVwwYILSAwcYGACG/ARbHXQf2wAAAABJRU5ErkJggg=="

// testDestination holds config for a single ZTNA destination entry.
type testDestination struct {
	destination string
	port        string
	protocol    string
	subtype     string
}

// testAppConfig holds all parameters for testAccApplicationConfig.
type testAppConfig struct {
	resourceName    string // HCL resource label, e.g. "test_web"
	name            string // application display name
	appType         string // "web", "saas", or "ztna"
	description     string
	url             string
	hidden          bool
	agentlessAccess bool
	mobileSecurity  bool
	sbsOnlyLaunch   bool
	usingTemplate   bool
	icon            string
	relatedURLs     []string
	keywords        []string
	sso             string // HCL value for the sso attribute, e.g. `{ type = "nosso" }`
	destinations    []testDestination
	dependsOn       []string // HCL resource addresses for depends_on, e.g. ["spa_routing_domain.foo"]
}

// testAccApplicationConfig generates a Terraform HCL config for a spa_application resource.
func testAccApplicationConfig(cfg testAppConfig) string {
	if cfg.icon == "" {
		cfg.icon = testAppIcon
	}
	if cfg.resourceName == "" {
		cfg.resourceName = "test"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "resource \"spa_application\" %q {\n", cfg.resourceName)
	fmt.Fprintf(&b, "  name             = %q\n", cfg.name)
	fmt.Fprintf(&b, "  type             = %q\n", cfg.appType)
	if cfg.description != "" {
		fmt.Fprintf(&b, "  description      = %q\n", cfg.description)
	}
	if cfg.url != "" {
		fmt.Fprintf(&b, "  url              = %q\n", cfg.url)
	}
	fmt.Fprintf(&b, "  hidden           = %v\n", cfg.hidden)
	fmt.Fprintf(&b, "  agentless_access = %v\n", cfg.agentlessAccess)
	fmt.Fprintf(&b, "  mobile_security  = %v\n", cfg.mobileSecurity)
	fmt.Fprintf(&b, "  sbs_only_launch  = %v\n", cfg.sbsOnlyLaunch)
	fmt.Fprintf(&b, "  using_template   = %v\n", cfg.usingTemplate)
	fmt.Fprintf(&b, "  icon             = %q\n", cfg.icon)
	if len(cfg.relatedURLs) > 0 {
		urls := make([]string, len(cfg.relatedURLs))
		for i, u := range cfg.relatedURLs {
			urls[i] = fmt.Sprintf("%q", u)
		}
		fmt.Fprintf(&b, "  related_urls     = [%s]\n", strings.Join(urls, ", "))
	}
	if len(cfg.keywords) > 0 {
		kws := make([]string, len(cfg.keywords))
		for i, k := range cfg.keywords {
			kws[i] = fmt.Sprintf("%q", k)
		}
		fmt.Fprintf(&b, "  keywords         = [%s]\n", strings.Join(kws, ", "))
	}
	if cfg.sso != "" {
		fmt.Fprintf(&b, "  sso              = %s\n", cfg.sso)
	}
	if len(cfg.destinations) > 0 {
		fmt.Fprintf(&b, "  destination = [\n")
		for i, d := range cfg.destinations {
			fmt.Fprintf(&b, "    {\n")
			fmt.Fprintf(&b, "      destination = %q\n", d.destination)
			fmt.Fprintf(&b, "      port        = %q\n", d.port)
			fmt.Fprintf(&b, "      protocol    = %q\n", d.protocol)
			fmt.Fprintf(&b, "      subtype     = %q\n", d.subtype)
			if i < len(cfg.destinations)-1 {
				fmt.Fprintf(&b, "    },\n")
			} else {
				fmt.Fprintf(&b, "    }\n")
			}
		}
		fmt.Fprintf(&b, "  ]\n")
	}
	if len(cfg.dependsOn) > 0 {
		fmt.Fprintf(&b, "  depends_on = [%s]\n", strings.Join(cfg.dependsOn, ", "))
	}
	fmt.Fprintf(&b, "}\n")
	return b.String()
}

func testAccCheckApplicationDestroy(s *terraform.State) error {
	client, err := testAccCreateClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "spa_application" {
			continue
		}

		id := rs.Primary.Attributes["id"]
		_, err := client.GetApplication(ctx, id)
		if err == nil {
			return fmt.Errorf("application %s still exists in the API after destroy", id)
		}
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("unexpected error checking application %s: %s", id, err)
		}
	}
	return nil
}

func testAccCheckApplicationExistsInAPI(resourceName string) resource.TestCheckFunc {
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

		app, err := client.GetApplication(context.Background(), id)
		if err != nil {
			return fmt.Errorf("application %s not found in API: %s", id, err)
		}

		if app.Name != rs.Primary.Attributes["name"] {
			return fmt.Errorf("application name mismatch: API=%q, state=%q", app.Name, rs.Primary.Attributes["name"])
		}

		return nil
	}
}

func TestAccApplication_web(t *testing.T) {
	name := "tf-acc-test-web-app"
	fqdn := fmt.Sprintf("%s.example.com", name)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccApplicationConfig(testAppConfig{
					resourceName: "test_web",
					name:         name,
					appType:      "web",
					description:  "Terraform acceptance test - web application",
					url:          fmt.Sprintf("https://%s", fqdn),
					relatedURLs:  []string{"*.example.com"},
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("spa_application.test_web"),
					resource.TestCheckResourceAttr("spa_application.test_web", "name", name),
					resource.TestCheckResourceAttr("spa_application.test_web", "type", "web"),
					resource.TestCheckResourceAttr("spa_application.test_web", "description", "Terraform acceptance test - web application"),
					resource.TestCheckResourceAttr("spa_application.test_web", "url", fmt.Sprintf("https://%s", fqdn)),
					resource.TestCheckResourceAttr("spa_application.test_web", "hidden", "false"),
					resource.TestCheckResourceAttr("spa_application.test_web", "agentless_access", "false"),
					resource.TestCheckResourceAttr("spa_application.test_web", "mobile_security", "false"),
					resource.TestCheckResourceAttr("spa_application.test_web", "sbs_only_launch", "false"),
					resource.TestCheckResourceAttr("spa_application.test_web", "using_template", "false"),
					resource.TestCheckResourceAttr("spa_application.test_web", "related_urls.#", "1"),
					resource.TestCheckTypeSetElemAttr("spa_application.test_web", "related_urls.*", "*.example.com"),
					resource.TestCheckResourceAttrSet("spa_application.test_web", "id"),
					resource.TestCheckResourceAttrSet("spa_application.test_web", "state"),
				),
			},
			// Step 2: Update — change description, add a keyword, set hidden=true
			{
				Config: testAccApplicationConfig(testAppConfig{
					resourceName: "test_web",
					name:         name,
					appType:      "web",
					description:  "Terraform acceptance test - web application UPDATED",
					url:          fmt.Sprintf("https://%s", fqdn),
					hidden:       true,
					relatedURLs:  []string{"*.example.com", fmt.Sprintf("api.%s", fqdn)},
					keywords:     []string{"acceptance-test"},
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("spa_application.test_web"),
					resource.TestCheckResourceAttr("spa_application.test_web", "name", name),
					resource.TestCheckResourceAttr("spa_application.test_web", "type", "web"),
					resource.TestCheckResourceAttr("spa_application.test_web", "description", "Terraform acceptance test - web application UPDATED"),
					resource.TestCheckResourceAttr("spa_application.test_web", "url", fmt.Sprintf("https://%s", fqdn)),
					resource.TestCheckResourceAttr("spa_application.test_web", "hidden", "true"),
					resource.TestCheckResourceAttr("spa_application.test_web", "agentless_access", "false"),
					resource.TestCheckResourceAttr("spa_application.test_web", "mobile_security", "false"),
					resource.TestCheckResourceAttr("spa_application.test_web", "sbs_only_launch", "false"),
					resource.TestCheckResourceAttr("spa_application.test_web", "using_template", "false"),
					resource.TestCheckResourceAttr("spa_application.test_web", "related_urls.#", "2"),
					resource.TestCheckTypeSetElemAttr("spa_application.test_web", "related_urls.*", "*.example.com"),
					resource.TestCheckTypeSetElemAttr("spa_application.test_web", "related_urls.*", fmt.Sprintf("api.%s", fqdn)),
					resource.TestCheckResourceAttr("spa_application.test_web", "keywords.#", "1"),
					resource.TestCheckTypeSetElemAttr("spa_application.test_web", "keywords.*", "acceptance-test"),
					resource.TestCheckResourceAttrSet("spa_application.test_web", "id"),
					resource.TestCheckResourceAttrSet("spa_application.test_web", "state"),
				),
			},
			// Step 3: ImportState — verify the resource can be imported by ID
			{
				ResourceName:      "spa_application.test_web",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccApplication_saas(t *testing.T) {
	name := "tf-acc-test-saas-app"
	fqdn := fmt.Sprintf("%s.example.com", name)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccApplicationConfig(testAppConfig{
					resourceName:    "test_saas",
					name:            name,
					appType:         "saas",
					description:     "Terraform acceptance test - SaaS application",
					url:             fmt.Sprintf("https://%s", fqdn),
					agentlessAccess: true,
					sbsOnlyLaunch:   true,
					relatedURLs:     []string{fmt.Sprintf("*.%s", fqdn)},
					keywords:        []string{"acceptance-test"},
					sso:             `{ type = "nosso" }`,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("spa_application.test_saas"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "name", name),
					resource.TestCheckResourceAttr("spa_application.test_saas", "type", "saas"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "description", "Terraform acceptance test - SaaS application"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "url", fmt.Sprintf("https://%s", fqdn)),
					resource.TestCheckResourceAttr("spa_application.test_saas", "hidden", "false"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "agentless_access", "true"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "mobile_security", "false"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "sbs_only_launch", "true"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "using_template", "false"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "related_urls.#", "1"),
					resource.TestCheckTypeSetElemAttr("spa_application.test_saas", "related_urls.*", fmt.Sprintf("*.%s", fqdn)),
					resource.TestCheckResourceAttr("spa_application.test_saas", "keywords.#", "1"),
					resource.TestCheckTypeSetElemAttr("spa_application.test_saas", "keywords.*", "acceptance-test"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "sso.type", "nosso"),
					resource.TestCheckResourceAttrSet("spa_application.test_saas", "id"),
					resource.TestCheckResourceAttrSet("spa_application.test_saas", "state"),
				),
			},
			// Step 2: Update — change description, add a keyword, set hidden=true, add a related URL
			{
				Config: testAccApplicationConfig(testAppConfig{
					resourceName:    "test_saas",
					name:            name,
					appType:         "saas",
					description:     "Terraform acceptance test - SaaS application UPDATED",
					url:             fmt.Sprintf("https://%s", fqdn),
					agentlessAccess: true,
					hidden:          true,
					sbsOnlyLaunch:   true,
					relatedURLs:     []string{fmt.Sprintf("*.%s", fqdn), fmt.Sprintf("api.%s", fqdn)},
					keywords:        []string{"acceptance-test", "updated"},
					sso:             `{ type = "nosso" }`,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("spa_application.test_saas"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "name", name),
					resource.TestCheckResourceAttr("spa_application.test_saas", "type", "saas"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "description", "Terraform acceptance test - SaaS application UPDATED"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "url", fmt.Sprintf("https://%s", fqdn)),
					resource.TestCheckResourceAttr("spa_application.test_saas", "hidden", "true"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "agentless_access", "true"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "mobile_security", "false"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "sbs_only_launch", "true"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "using_template", "false"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "related_urls.#", "2"),
					resource.TestCheckTypeSetElemAttr("spa_application.test_saas", "related_urls.*", fmt.Sprintf("*.%s", fqdn)),
					resource.TestCheckTypeSetElemAttr("spa_application.test_saas", "related_urls.*", fmt.Sprintf("api.%s", fqdn)),
					resource.TestCheckResourceAttr("spa_application.test_saas", "keywords.#", "2"),
					resource.TestCheckTypeSetElemAttr("spa_application.test_saas", "keywords.*", "acceptance-test"),
					resource.TestCheckTypeSetElemAttr("spa_application.test_saas", "keywords.*", "updated"),
					resource.TestCheckResourceAttr("spa_application.test_saas", "sso.type", "nosso"),
					resource.TestCheckResourceAttrSet("spa_application.test_saas", "id"),
					resource.TestCheckResourceAttrSet("spa_application.test_saas", "state"),
				),
			},
			// Step 3: ImportState — verify the resource can be imported by ID
			{
				ResourceName:      "spa_application.test_saas",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccApplication_ztna(t *testing.T) {
	name := "tf-acc-test-ztna-app"
	fqdn := fmt.Sprintf("%s.internal.example.com", name)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: testAccApplicationConfig(testAppConfig{
					resourceName: "test_ztna",
					name:         name,
					appType:      "ztna",
					description:  "Terraform acceptance test - ZTNA application",
					destinations: []testDestination{
						{
							destination: fqdn,
							port:        "443",
							protocol:    "PROTOCOL_TCP",
							subtype:     "SUBTYPE_HOSTNAME",
						},
					},
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("spa_application.test_ztna"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "name", name),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "type", "ztna"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "description", "Terraform acceptance test - ZTNA application"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "hidden", "false"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "agentless_access", "false"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "mobile_security", "false"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "sbs_only_launch", "false"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "using_template", "false"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "destination.#", "1"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "destination.0.destination", fqdn),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "destination.0.port", "443"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "destination.0.protocol", "PROTOCOL_TCP"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "destination.0.subtype", "SUBTYPE_HOSTNAME"),
					resource.TestCheckResourceAttrSet("spa_application.test_ztna", "id"),
					resource.TestCheckResourceAttrSet("spa_application.test_ztna", "state"),
				),
			},
			// Step 2: Update — change description and add a second destination
			{
				Config: testAccApplicationConfig(testAppConfig{
					resourceName: "test_ztna",
					name:         name,
					appType:      "ztna",
					description:  "Terraform acceptance test - ZTNA application UPDATED",
					destinations: []testDestination{
						{
							destination: fqdn,
							port:        "443",
							protocol:    "PROTOCOL_TCP",
							subtype:     "SUBTYPE_HOSTNAME",
						},
						{
							destination: fqdn,
							port:        "8443",
							protocol:    "PROTOCOL_TCP",
							subtype:     "SUBTYPE_HOSTNAME",
						},
					},
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("spa_application.test_ztna"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "name", name),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "type", "ztna"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "description", "Terraform acceptance test - ZTNA application UPDATED"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "hidden", "false"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "agentless_access", "false"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "mobile_security", "false"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "sbs_only_launch", "false"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "using_template", "false"),
					resource.TestCheckResourceAttr("spa_application.test_ztna", "destination.#", "2"),
					resource.TestCheckResourceAttrSet("spa_application.test_ztna", "id"),
					resource.TestCheckResourceAttrSet("spa_application.test_ztna", "state"),
				),
			},
			// Step 3: ImportState — verify the resource can be imported by ID
			{
				ResourceName:      "spa_application.test_ztna",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
