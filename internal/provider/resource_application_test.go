package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// =============================================================================
// Application Tests
// =============================================================================

// TestShouldAwaitSSO asserts the Read gate: use the bounded SSO re-fetch on
// import (prior name null) or when prior state already had an SSO object, and a
// single GET for a settled application with no SSO.
func TestShouldAwaitSSO(t *testing.T) {
	nonNullSSO := types.ObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{})
	nullSSO := types.ObjectNull(map[string]attr.Type{})

	cases := []struct {
		desc string
		name types.String
		sso  types.Object
		want bool
	}{
		{"import: id-only prior state (name null)", types.StringNull(), nullSSO, true},
		{"import with prior SSO present", types.StringNull(), nonNullSSO, true},
		{"refresh: prior state has SSO", types.StringValue("app"), nonNullSSO, true},
		{"refresh: settled app with no SSO", types.StringValue("app"), nullSSO, false},
	}
	for _, tc := range cases {
		if got := shouldAwaitSSO(tc.name, tc.sso); got != tc.want {
			t.Errorf("%s: shouldAwaitSSO=%v, want %v", tc.desc, got, tc.want)
		}
	}
}

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
	resourceName      string // HCL resource label, e.g. "test_web"
	name              string // application display name
	appType           string // "web", "saas", or "ztna"
	description       string
	url               string
	hidden            bool
	agentlessAccess   bool
	mobileSecurity    bool
	sbsOnlyLaunch     bool
	usingTemplate     bool
	omitUsingTemplate bool // when true, omit using_template entirely (tests the Computed default)
	templateName      string
	icon              string
	relatedURLs       []string
	keywords          []string
	sso               string // HCL value for the sso attribute, e.g. `{ type = "nosso" }`
	state             string // "incomplete" or "complete"; omitted if empty
	destinations      []testDestination
	dependsOn         []string // HCL resource addresses for depends_on, e.g. ["citrixspa_routing_domain.foo"]
}

// testAccApplicationConfig generates a Terraform HCL config for a citrixspa_application resource.
func testAccApplicationConfig(cfg testAppConfig) string {
	if cfg.icon == "" {
		cfg.icon = testAppIcon
	}
	if cfg.resourceName == "" {
		cfg.resourceName = "test"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "resource \"citrixspa_application\" %q {\n", cfg.resourceName)
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
	if !cfg.omitUsingTemplate {
		fmt.Fprintf(&b, "  using_template   = %v\n", cfg.usingTemplate)
	}
	if cfg.templateName != "" {
		fmt.Fprintf(&b, "  template_name    = %q\n", cfg.templateName)
	}
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
	if cfg.state != "" {
		fmt.Fprintf(&b, "  state            = %q\n", cfg.state)
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
		if rs.Type != "citrixspa_application" {
			continue
		}

		id := rs.Primary.Attributes["id"]
		_, err := client.GetApplication(ctx, id)
		if err == nil {
			return fmt.Errorf("application %s still exists in the API after destroy", id)
		}
		if !IsNotFound(err) {
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
	// Shared create config, reused by the create and idempotency steps so they
	// stay identical (the idempotency step must re-apply the SAME config).
	createConfig := testAccApplicationConfig(testAppConfig{
		resourceName: "test_web",
		name:         name,
		appType:      "web",
		description:  "Terraform acceptance test - web application",
		url:          fmt.Sprintf("https://%s", fqdn),
		relatedURLs:  []string{"*.example.com"},
	})
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create
			{
				Config: createConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("citrixspa_application.test_web"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "name", name),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "type", "web"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "description", "Terraform acceptance test - web application"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "url", fmt.Sprintf("https://%s", fqdn)),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "hidden", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "agentless_access", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "mobile_security", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "sbs_only_launch", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "using_template", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "related_urls.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrixspa_application.test_web", "related_urls.*", "*.example.com"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_web", "id"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_web", "state"),
				),
			},
			// Step 2: Idempotency — re-applying the create config yields an empty plan
			{
				Config:   createConfig,
				PlanOnly: true,
			},
			// Step 3: Update — change description, add a keyword, set hidden=true
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
					testAccCheckApplicationExistsInAPI("citrixspa_application.test_web"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "name", name),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "type", "web"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "description", "Terraform acceptance test - web application UPDATED"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "url", fmt.Sprintf("https://%s", fqdn)),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "hidden", "true"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "agentless_access", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "mobile_security", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "sbs_only_launch", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "using_template", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "related_urls.#", "2"),
					resource.TestCheckTypeSetElemAttr("citrixspa_application.test_web", "related_urls.*", "*.example.com"),
					resource.TestCheckTypeSetElemAttr("citrixspa_application.test_web", "related_urls.*", fmt.Sprintf("api.%s", fqdn)),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "keywords.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrixspa_application.test_web", "keywords.*", "acceptance-test"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_web", "id"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_web", "state"),
				),
			},
			// Step 4: Omit description — the backend enforces a minimum length and
			// retains the prior value when the field is omitted, so a set
			// description cannot be cleared. Omitting it keeps the step-3 value
			// (UseStateForUnknown) with an empty plan and no inconsistent-result error.
			{
				Config: testAccApplicationConfig(testAppConfig{
					resourceName: "test_web",
					name:         name,
					appType:      "web",
					url:          fmt.Sprintf("https://%s", fqdn),
					hidden:       true,
					relatedURLs:  []string{"*.example.com", fmt.Sprintf("api.%s", fqdn)},
					keywords:     []string{"acceptance-test"},
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("citrixspa_application.test_web"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web", "description", "Terraform acceptance test - web application UPDATED"),
				),
			},
			// Step 5: ImportState — verify the resource can be imported by ID
			{
				ResourceName:      "citrixspa_application.test_web",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccApplication_usingTemplateOmitted is a regression test for the
// Console-owned `using_template` flag. Because the Console owns/normalizes this
// field, the provider declares it Optional + Computed. Omitting it from config
// must NOT fail with "Provider produced inconsistent result after apply"
// (.using_template: was null, but now cty.False), and a re-apply of the same
// config must produce an empty plan (no spurious drift).
func TestAccApplication_usingTemplateOmitted(t *testing.T) {
	name := "tf-acc-test-using-template-omitted"
	fqdn := fmt.Sprintf("%s.example.com", name)
	// using_template is intentionally omitted from the config.
	config := fmt.Sprintf(`
resource "citrixspa_application" "test_ut" {
  name         = %[1]q
  type         = "web"
  description  = "Terraform acceptance test - using_template omitted"
  url          = "https://%[2]s"
  icon         = %[3]q
  related_urls = ["*.example.com"]
}
`, name, fqdn, testAppIcon)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create — the computed value must be populated.
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("citrixspa_application.test_ut"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_ut", "using_template"),
				),
			},
			// Step 2: Idempotency — re-applying the same config yields no drift.
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

// TestAccApplication_webNossoSSO is a regression test for a web application
// created with an explicit sso = { type = "nosso" } block.
//
// The backend defaults a web app's SSO to {type:"nosso"} but omits the SSO
// object from the GET that immediately follows creation, so the post-create
// read returns no SSO. When the configuration explicitly set the SSO, returning
// null violated Terraform's plan==result contract and surfaced as "Provider
// produced inconsistent result after apply". The create path now restores the
// configured SSO when the read-back omits it. On the pre-fix binary this step
// deterministically errors on create; the idempotency step then proves state
// matches the config.
func TestAccApplication_webNossoSSO(t *testing.T) {
	name := "tf-acc-test-web-nosso-app"
	fqdn := fmt.Sprintf("%s.example.com", name)
	createConfig := testAccApplicationConfig(testAppConfig{
		resourceName: "test_web_nosso",
		name:         name,
		appType:      "web",
		description:  "Terraform acceptance test - web application with explicit nosso SSO",
		url:          fmt.Sprintf("https://%s", fqdn),
		relatedURLs:  []string{"*.example.com"},
		sso:          `{ type = "nosso" }`,
	})
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create — must apply cleanly (no inconsistent-result error).
			{
				Config: createConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("citrixspa_application.test_web_nosso"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web_nosso", "name", name),
					resource.TestCheckResourceAttr("citrixspa_application.test_web_nosso", "type", "web"),
					resource.TestCheckResourceAttr("citrixspa_application.test_web_nosso", "sso.type", "nosso"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_web_nosso", "id"),
				),
			},
			// Step 2: Idempotency — re-applying the same config yields an empty plan.
			{
				Config:   createConfig,
				PlanOnly: true,
			},
			// Step 3: ImportState — verify the resource can be imported by ID.
			{
				ResourceName:      "citrixspa_application.test_web_nosso",
				ImportState:       true,
				ImportStateVerify: true,
			},
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
					testAccCheckApplicationExistsInAPI("citrixspa_application.test_saas"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "name", name),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "type", "saas"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "description", "Terraform acceptance test - SaaS application"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "url", fmt.Sprintf("https://%s", fqdn)),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "hidden", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "agentless_access", "true"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "mobile_security", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "sbs_only_launch", "true"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "using_template", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "related_urls.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrixspa_application.test_saas", "related_urls.*", fmt.Sprintf("*.%s", fqdn)),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "keywords.#", "1"),
					resource.TestCheckTypeSetElemAttr("citrixspa_application.test_saas", "keywords.*", "acceptance-test"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "sso.type", "nosso"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_saas", "id"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_saas", "state"),
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
					testAccCheckApplicationExistsInAPI("citrixspa_application.test_saas"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "name", name),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "type", "saas"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "description", "Terraform acceptance test - SaaS application UPDATED"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "url", fmt.Sprintf("https://%s", fqdn)),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "hidden", "true"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "agentless_access", "true"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "mobile_security", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "sbs_only_launch", "true"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "using_template", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "related_urls.#", "2"),
					resource.TestCheckTypeSetElemAttr("citrixspa_application.test_saas", "related_urls.*", fmt.Sprintf("*.%s", fqdn)),
					resource.TestCheckTypeSetElemAttr("citrixspa_application.test_saas", "related_urls.*", fmt.Sprintf("api.%s", fqdn)),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "keywords.#", "2"),
					resource.TestCheckTypeSetElemAttr("citrixspa_application.test_saas", "keywords.*", "acceptance-test"),
					resource.TestCheckTypeSetElemAttr("citrixspa_application.test_saas", "keywords.*", "updated"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saas", "sso.type", "nosso"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_saas", "id"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_saas", "state"),
				),
			},
			// Step 3: ImportState — verify the resource can be imported by ID
			{
				ResourceName:      "citrixspa_application.test_saas",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccApplication_iconNormalization guards the icon custom-type wiring end to
// end: it applies an icon carrying a data-URI prefix and a newline inside the
// base64 payload. The SPA service stores the icon stripped of both, so unless
// the framework builds an IconValue and runs semantic equality on the refresh,
// the read-back value differs from config and apply fails with "inconsistent
// result after apply". The PlanOnly step asserts an empty follow-up plan, so a
// wiring regression cannot silently reintroduce that failure.
func TestAccApplication_iconNormalization(t *testing.T) {
	name := "tf-acc-test-icon-norm"
	fqdn := fmt.Sprintf("%s.example.com", name)
	wrappedIcon := "data:image/png;base64," + testAppIcon[:40] + "\n" + testAppIcon[40:]
	cfg := testAccApplicationConfig(testAppConfig{
		resourceName: "test_icon",
		name:         name,
		appType:      "saas",
		description:  "Terraform acceptance test - icon normalization",
		url:          fmt.Sprintf("https://%s", fqdn),
		relatedURLs:  []string{fmt.Sprintf("*.%s", fqdn)},
		sso:          `{ type = "nosso" }`,
		icon:         wrappedIcon,
	})
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with a data-URI + newline-wrapped icon
			{
				Config: cfg,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("citrixspa_application.test_icon"),
					resource.TestCheckResourceAttr("citrixspa_application.test_icon", "name", name),
					resource.TestCheckResourceAttr("citrixspa_application.test_icon", "icon", wrappedIcon),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_icon", "id"),
				),
			},
			// Step 2: Idempotency — the same wrapped icon must yield an empty plan
			{
				Config:   cfg,
				PlanOnly: true,
			},
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
					testAccCheckApplicationExistsInAPI("citrixspa_application.test_ztna"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "name", name),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "type", "ztna"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "description", "Terraform acceptance test - ZTNA application"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "hidden", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "agentless_access", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "mobile_security", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "sbs_only_launch", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "using_template", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "destination.#", "1"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "destination.0.destination", fqdn),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "destination.0.port", "443"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "destination.0.protocol", "PROTOCOL_TCP"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "destination.0.subtype", "SUBTYPE_HOSTNAME"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_ztna", "id"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_ztna", "state"),
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
					testAccCheckApplicationExistsInAPI("citrixspa_application.test_ztna"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "name", name),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "type", "ztna"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "description", "Terraform acceptance test - ZTNA application UPDATED"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "hidden", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "agentless_access", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "mobile_security", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "sbs_only_launch", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "using_template", "false"),
					resource.TestCheckResourceAttr("citrixspa_application.test_ztna", "destination.#", "2"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_ztna", "id"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_ztna", "state"),
				),
			},
			// Step 3: ImportState — verify the resource can be imported by ID
			{
				ResourceName:      "citrixspa_application.test_ztna",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccApplication_samlSSO verifies that SAML SSO applications can be created
// and updated without "Provider produced inconsistent result after apply" errors.
// Covers:
// - Server-computed SSO fields (saml_sso_login_url) are stripped on create/update
// - custom_attributes returned as JSON string are parsed correctly
// - Null SSO values produce type-appropriate defaults
// - SSO object shape is preserved across plan/apply cycles
func TestAccApplication_samlSSO(t *testing.T) {
	name := "tf-acc-test-saml-sso-app"
	fqdn := fmt.Sprintf("%s.example.com", name)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with SAML SSO including writable fields only
			{
				Config: testAccApplicationConfig(testAppConfig{
					resourceName: "test_saml",
					name:         name,
					appType:      "saas",
					description:  "Terraform acceptance test - SAML SSO application",
					url:          fmt.Sprintf("https://%s", fqdn),
					relatedURLs:  []string{fmt.Sprintf("*.%s", fqdn)},
					sso: `{
						type              = "saml"
						assertion_url     = "https://sp.example.com/acs"
						audience          = "https://sp.example.com"
						name_id_format    = "emailAddress"
						name_id_source    = "email"
						custom_attributes = []
					}`,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("citrixspa_application.test_saml"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saml", "name", name),
					resource.TestCheckResourceAttr("citrixspa_application.test_saml", "type", "saas"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saml", "sso.type", "saml"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saml", "sso.assertion_url", "https://sp.example.com/acs"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saml", "sso.audience", "https://sp.example.com"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saml", "sso.name_id_format", "emailAddress"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saml", "sso.name_id_source", "email"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_saml", "id"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_saml", "state"),
				),
			},
			// Step 2: Update — change assertion_url
			{
				Config: testAccApplicationConfig(testAppConfig{
					resourceName: "test_saml",
					name:         name,
					appType:      "saas",
					description:  "Terraform acceptance test - SAML SSO application UPDATED",
					url:          fmt.Sprintf("https://%s", fqdn),
					relatedURLs:  []string{fmt.Sprintf("*.%s", fqdn)},
					sso: `{
						type              = "saml"
						assertion_url     = "https://sp.example.com/acs/v2"
						audience          = "https://sp.example.com"
						name_id_format    = "emailAddress"
						name_id_source    = "email"
						custom_attributes = []
					}`,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("citrixspa_application.test_saml"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saml", "name", name),
					resource.TestCheckResourceAttr("citrixspa_application.test_saml", "description", "Terraform acceptance test - SAML SSO application UPDATED"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saml", "sso.type", "saml"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saml", "sso.assertion_url", "https://sp.example.com/acs/v2"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_saml", "id"),
				),
			},
			// Step 3: ImportState — with SingleNestedAttribute the SSO shape is
			// fixed, so import now works without ignoring the sso field.
			{
				ResourceName:      "citrixspa_application.test_saml",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccApplication_samlSSOComputedFieldsPopulated verifies that server-computed
// SSO fields (saml_sso_login_url, saml_cert_issuer_name) are populated by the
// server after create, without the user needing to set them in config.
func TestAccApplication_samlSSOComputedFieldsPopulated(t *testing.T) {
	name := "tf-acc-test-saml-computed"
	fqdn := fmt.Sprintf("%s.example.com", name)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			// Create without computed fields — they should be populated by the server
			{
				Config: testAccApplicationConfig(testAppConfig{
					resourceName: "test_saml_computed",
					name:         name,
					appType:      "saas",
					description:  "Terraform acceptance test - SAML SSO computed fields populated",
					url:          fmt.Sprintf("https://%s", fqdn),
					relatedURLs:  []string{fmt.Sprintf("*.%s", fqdn)},
					sso: `{
						type              = "saml"
						assertion_url     = "https://sp.example.com/acs"
						audience          = "https://sp.example.com"
						name_id_format    = "emailAddress"
						name_id_source    = "email"
						custom_attributes = []
					}`,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("citrixspa_application.test_saml_computed"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saml_computed", "name", name),
					resource.TestCheckResourceAttr("citrixspa_application.test_saml_computed", "sso.type", "saml"),
					resource.TestCheckResourceAttr("citrixspa_application.test_saml_computed", "sso.assertion_url", "https://sp.example.com/acs"),
					// Server-computed fields should be populated
					resource.TestCheckResourceAttrSet("citrixspa_application.test_saml_computed", "sso.saml_sso_login_url"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_saml_computed", "id"),
				),
			},
		},
	})
}

// TestAccApplication_template exercises the "provision from catalog template"
// path used by the shipped templates/ files: an application created with
// using_template = true and a template_name, alongside its routing domain and a
// depends_on edge. It mirrors the shape of the generated SaaS templates but uses
// concrete (non-placeholder) URLs so it applies against a live tenant.
func TestAccApplication_template(t *testing.T) {
	name := "tf-acc-test-template-app"
	rdHost := "tf-acc-test-template.example.com"
	rdFQDN := "*." + rdHost
	config := testAccRoutingDomainConfig("rd_tmpl", rdFQDN, "external", "saas", "Terraform acceptance test - template", "enabled", "false", "[]") +
		testAccApplicationConfig(testAppConfig{
			resourceName:  "test_template",
			name:          name,
			appType:       "saas",
			description:   "Terraform acceptance test - template application",
			url:           fmt.Sprintf("https://%s", rdHost),
			usingTemplate: true,
			templateName:  "Salesforce",
			relatedURLs:   []string{rdFQDN},
			sso: `{
				type              = "saml"
				assertion_url     = "https://sp.example.com/acs"
				audience          = "https://sp.example.com"
				sign_assertion    = "ASSERTION"
				name_id_source    = "name"
				name_id_format    = "transient"
				saml_type         = "SP_IDP"
				sp_initiated_only = false
			}`,
			dependsOn: []string{"citrixspa_routing_domain.rd_tmpl"},
		})
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupRoutingDomain(rdFQDN)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckApplicationDestroy,
			testAccCheckRoutingDomainDestroy,
		),
		Steps: []resource.TestStep{
			// Step 1: Create from template
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExistsInAPI("citrixspa_application.test_template"),
					testAccCheckRoutingDomainExistsInAPI("citrixspa_routing_domain.rd_tmpl"),
					resource.TestCheckResourceAttr("citrixspa_application.test_template", "name", name),
					resource.TestCheckResourceAttr("citrixspa_application.test_template", "type", "saas"),
					resource.TestCheckResourceAttr("citrixspa_application.test_template", "using_template", "true"),
					resource.TestCheckResourceAttr("citrixspa_application.test_template", "template_name", "Salesforce"),
					resource.TestCheckResourceAttr("citrixspa_application.test_template", "sso.type", "saml"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_template", "id"),
					resource.TestCheckResourceAttrSet("citrixspa_application.test_template", "state"),
				),
			},
			// Step 2: Idempotency — re-applying the template config yields an empty plan
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

// =============================================================================
// Unit tests — FQDN derivation and the routing-domain warning emitted on delete
// =============================================================================

var testDestinationObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"destination": types.StringType,
		"port":        types.StringType,
		"protocol":    types.StringType,
		"subtype":     types.StringType,
	},
}

var testLocationObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name": types.StringType,
		"uuid": types.StringType,
	},
}

var testPolicyObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"type": types.StringType,
		"data": types.MapType{ElemType: types.StringType},
	},
}

func testStringSet(values ...string) types.Set {
	elements := make([]attr.Value, 0, len(values))
	for _, value := range values {
		elements = append(elements, types.StringValue(value))
	}
	return types.SetValueMust(types.StringType, elements)
}

func testDestinationList(destinations ...string) types.List {
	elements := make([]attr.Value, 0, len(destinations))
	for _, destination := range destinations {
		elements = append(elements, types.ObjectValueMust(testDestinationObjectType.AttrTypes, map[string]attr.Value{
			"destination": types.StringValue(destination),
			"port":        types.StringValue("443"),
			"protocol":    types.StringValue("PROTOCOL_TCP"),
			"subtype":     types.StringValue("SUBTYPE_HOSTNAME"),
		}))
	}
	return types.ListValueMust(testDestinationObjectType, elements)
}

func TestApplicationFQDNs(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		desc        string
		url         types.String
		relatedURLs types.Set
		destination types.List
		want        []string
	}{
		{
			// net/url reads "host:port" without a scheme as scheme plus opaque
			// data, so the hostname is empty unless the value is re-parsed as an
			// authority. The backend does that, so the provider must too.
			desc:        "scheme-less url with a port reduces to the hostname",
			url:         types.StringValue("intranet.example.com:8443"),
			relatedURLs: types.SetNull(types.StringType),
			destination: types.ListNull(testDestinationObjectType),
			want:        []string{"intranet.example.com"},
		},
		{
			desc:        "url hostname is lower-cased to match what the backend stores",
			url:         types.StringValue("https://Intranet.EXAMPLE.com/app"),
			relatedURLs: types.SetNull(types.StringType),
			destination: types.ListNull(testDestinationObjectType),
			want:        []string{"intranet.example.com"},
		},
		{
			// The service stores any url without a bare "%" or a double quote, so
			// a value Go cannot parse reaches state and reaches this warning. Only
			// the authority may be reported: a path or query can carry a token.
			desc:        "url that cannot be parsed reports only its authority",
			url:         types.StringValue("ht tp://example.com/app?token=notarealsecret"),
			relatedURLs: types.SetNull(types.StringType),
			destination: types.ListNull(testDestinationObjectType),
			want:        []string{"example.com"},
		},
		{
			desc:        "unparseable related url does not echo its query",
			url:         types.StringNull(),
			relatedURLs: testStringSet("ht tp://api.example.com/v1?apikey=notarealsecret"),
			destination: types.ListNull(testDestinationObjectType),
			want:        []string{"api.example.com"},
		},
		{
			// ZTNA destinations are stored verbatim, so they must not be parsed:
			// url.Parse("//10.60.0.0/24") yields the host 10.60.0.0 and silently
			// widens the reported range.
			desc:        "ztna cidr destination keeps its suffix",
			url:         types.StringNull(),
			relatedURLs: types.SetNull(types.StringType),
			destination: testDestinationList("10.60.0.0/24"),
			want:        []string{"10.60.0.0/24"},
		},
		{
			desc:        "scheme-less url with embedded credentials drops the userinfo",
			url:         types.StringValue("user:pass@example.com"),
			relatedURLs: types.SetNull(types.StringType),
			destination: types.ListNull(testDestinationObjectType),
			want:        []string{"example.com"},
		},
		{
			desc:        "credentials in a related url are not echoed back",
			url:         types.StringNull(),
			relatedURLs: testStringSet("admin:secret@api.example.com"),
			destination: types.ListNull(testDestinationObjectType),
			want:        []string{"api.example.com"},
		},
		{
			desc:        "scheme-ful url is reduced to its bare hostname",
			url:         types.StringValue("https://example.com"),
			relatedURLs: types.SetNull(types.StringType),
			destination: types.ListNull(testDestinationObjectType),
			want:        []string{"example.com"},
		},
		{
			desc:        "url with a scheme and a port keeps only the hostname",
			url:         types.StringValue("https://example.com:8443"),
			relatedURLs: types.SetNull(types.StringType),
			destination: types.ListNull(testDestinationObjectType),
			want:        []string{"example.com"},
		},
		{
			desc:        "url and related urls are combined",
			url:         types.StringValue("https://example.com"),
			relatedURLs: testStringSet("api.example.com"),
			destination: types.ListNull(testDestinationObjectType),
			want:        []string{"example.com", "api.example.com"},
		},
		{
			desc:        "ztna destinations are emitted as configured",
			url:         types.StringNull(),
			relatedURLs: types.SetNull(types.StringType),
			destination: testDestinationList("database.internal.com", "10.0.1.0/24"),
			want:        []string{"database.internal.com", "10.0.1.0/24"},
		},
		{
			desc:        "null url and null collections yield nothing",
			url:         types.StringNull(),
			relatedURLs: types.SetNull(types.StringType),
			destination: types.ListNull(testDestinationObjectType),
			want:        []string{},
		},
		{
			desc:        "empty url is skipped entirely",
			url:         types.StringValue("   "),
			relatedURLs: testStringSet("api.example.com"),
			destination: types.ListNull(testDestinationObjectType),
			want:        []string{"api.example.com"},
		},
		{
			desc:        "unparseable url falls back to its authority",
			url:         types.StringValue("  ://missing-scheme  "),
			relatedURLs: types.SetNull(types.StringType),
			destination: types.ListNull(testDestinationObjectType),
			want:        []string{"missing-scheme"},
		},
		{
			desc:        "scheme-less url falls back to the raw trimmed value",
			url:         types.StringValue("example.com"),
			relatedURLs: types.SetNull(types.StringType),
			destination: types.ListNull(testDestinationObjectType),
			want:        []string{"example.com"},
		},
		{
			desc:        "a value repeated across url and related urls is emitted once",
			url:         types.StringValue("https://example.com"),
			relatedURLs: testStringSet("example.com"),
			destination: types.ListNull(testDestinationObjectType),
			want:        []string{"example.com"},
		},
	}

	for _, testCase := range testCases {
		data := ApplicationResourceModel{
			URL:         testCase.url,
			RelatedURLs: testCase.relatedURLs,
			Destination: testCase.destination,
		}

		got := applicationFQDNs(ctx, &data)

		if len(got) != len(testCase.want) {
			t.Errorf("%s: got %v, want %v", testCase.desc, got, testCase.want)
			continue
		}
		for i := range testCase.want {
			if got[i] != testCase.want[i] {
				t.Errorf("%s: got %v, want %v", testCase.desc, got, testCase.want)
				break
			}
		}
	}
}

type mockApplicationClient struct {
	SPAClient
	deleteApplication func(ctx context.Context, id string) error
}

func (m *mockApplicationClient) DeleteApplication(ctx context.Context, id string) error {
	return m.deleteApplication(ctx, id)
}

func testApplicationState() ApplicationResourceModel {
	return ApplicationResourceModel{
		ID:                   types.StringValue("00000000-0000-0000-0000-000000000000"),
		Name:                 types.StringValue("My Web Application"),
		Type:                 types.StringValue("web"),
		Description:          types.StringValue("A sample web application"),
		URL:                  types.StringValue("https://example.com"),
		Category:             types.StringValue("Productivity"),
		Hidden:               types.BoolValue(false),
		AgentlessAccess:      types.BoolValue(false),
		MobileSecurity:       types.BoolValue(false),
		SbsOnlyLaunch:        types.BoolValue(false),
		UsingTemplate:        types.BoolValue(false),
		TemplateName:         types.StringNull(),
		Icon:                 NewIconNull(),
		IconURL:              types.StringNull(),
		RelatedURLs:          testStringSet("api.example.com"),
		Keywords:             types.SetNull(types.StringType),
		Locations:            types.ListNull(testLocationObjectType),
		Policies:             types.ListNull(testPolicyObjectType),
		Destination:          types.ListNull(testDestinationObjectType),
		CustomProperties:     types.MapNull(types.StringType),
		CustomerDomainFields: types.MapNull(types.StringType),
		SSO:                  types.ObjectNull(ssoAttrTypes),
		State:                types.StringValue("complete"),
		PolicyCount:          types.StringValue("0"),
	}
}

func runApplicationDelete(t *testing.T, client SPAClient, model ApplicationResourceModel) diag.Diagnostics {
	t.Helper()
	ctx := context.Background()
	r := &ApplicationResource{client: client}

	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("failed to build schema: %v", schemaResp.Diagnostics.Errors())
	}

	state := tfsdk.State{Schema: schemaResp.Schema}
	if diags := state.Set(ctx, &model); diags.HasError() {
		t.Fatalf("failed to seed prior state: %v", diags.Errors())
	}

	deleteResp := &fwresource.DeleteResponse{State: state}
	r.Delete(ctx, fwresource.DeleteRequest{State: state}, deleteResp)

	return deleteResp.Diagnostics
}

func TestApplicationDeleteWarnsAboutMatchingRoutingDomains(t *testing.T) {
	client := &mockApplicationClient{
		deleteApplication: func(ctx context.Context, id string) error {
			return nil
		},
	}

	diags := runApplicationDelete(t, client, testApplicationState())

	if diags.HasError() {
		t.Fatalf("unexpected error diagnostics on a successful delete: %v", diags.Errors())
	}

	warnings := diags.Warnings()
	if len(warnings) != 1 {
		t.Fatalf("expected exactly one warning on delete, got %d: %v", len(warnings), warnings)
	}

	detail := warnings[0].Detail()

	for _, want := range []string{"example.com", "api.example.com", "Note 3"} {
		if !strings.Contains(detail, want) {
			t.Errorf("warning detail is missing %q: %q", want, detail)
		}
	}
	if strings.Contains(detail, "https://example.com") {
		t.Errorf("warning must list bare fqdns, not the scheme-ful url: %q", detail)
	}
	if strings.Contains(detail, "terraform apply") || strings.Contains(detail, "remove the resource block") {
		t.Errorf("the application delete warning must not carry a drift remedy: %q", detail)
	}
}

func TestApplicationDeleteDoesNotWarnWhenDeleteFails(t *testing.T) {
	client := &mockApplicationClient{
		deleteApplication: func(ctx context.Context, id string) error {
			return fmt.Errorf("API request failed with status 500: internal server error")
		},
	}

	diags := runApplicationDelete(t, client, testApplicationState())

	if !diags.HasError() {
		t.Fatalf("expected an error diagnostic when the delete call fails")
	}
	if got := len(diags.Warnings()); got != 0 {
		t.Errorf("expected no routing-domain warning when the delete fails, got %d: %v", got, diags.Warnings())
	}
}
