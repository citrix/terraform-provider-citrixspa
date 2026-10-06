package provider

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestResourceListingToolSelfTest runs the PowerShell discovery tool's own
// offline test suite (`spa_manager.ps1 -Test`).
//
// The suite includes the -ExtractLocals golden-file comparison, which is the
// only check that proves the flag still collapses shared users and resource
// locations the way it is supposed to. Hanging it off `go test` means it rides
// `make test`, `make check`, the e2e `tests` scenario and `release.yml` without
// any of them needing to know it exists.
//
// It needs no credentials and touches no tenant: everything is generated from
// resource-listing-tool/testdata/shared-values-dataset.json into a temp
// directory.
//
// Regenerate the golden files after an intentional output change:
//
//	SPA_UPDATE_GOLDEN=1 pwsh ./resource-listing-tool/spa_manager.ps1 -Test
func TestResourceListingToolSelfTest(t *testing.T) {
	// pwsh 7 is present on ubuntu-latest and is a hard requirement of the e2e
	// tooling job, so a missing pwsh here means a developer machine or a
	// community fork of the published repo — not a broken pipeline. Skipping
	// keeps `go test ./...` green for them; if the CI image ever loses pwsh,
	// e2e/tooling/run-spa-manager-roundtrip.sh fails loudly instead.
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not installed; skipping resource-listing-tool self-test")
	}

	script, err := filepath.Abs(filepath.Join("..", "..", "resource-listing-tool", "spa_manager.ps1"))
	if err != nil {
		t.Fatalf("resolving spa_manager.ps1 path: %v", err)
	}

	cmd := exec.Command("pwsh", "-NoProfile", "-File", script, "-Test")

	// Clear SPA_UPDATE_GOLDEN in the child. The suite's golden cases treat it as
	// "rewrite the expected files and skip the comparison", so a value inherited
	// from the developer's shell would make this test pass without comparing
	// anything AND silently rewrite the tracked testdata — a real regression
	// would be committed as the new expectation. Regenerating is a deliberate
	// act: run the pwsh command in the doc comment above, then read the diff.
	cmd.Env = append(os.Environ(), "SPA_UPDATE_GOLDEN=")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("spa_manager.ps1 -Test failed: %v\n%s", err, out)
	}
}
