package provider

import (
	"context"
	"testing"
)

// Regression: an explicit using_template must discriminate during orphan recovery,
// while an omitted using_template stays permissive (Console may flip it true).
func TestApplicationMatchesPlanned_usingTemplateDiscriminator(t *testing.T) {
	r := &ApplicationResource{}
	candidateTemplate := &ApplicationListItem{Name: "app", Type: "web", UsingTemplate: true}
	plannedFalse := &Application{Name: "app", Type: "web", UsingTemplate: false}

	// Explicit using_template = false vs candidate true → must NOT match.
	if r.applicationMatchesPlanned(context.Background(), candidateTemplate, plannedFalse, true) {
		t.Error("explicit using_template=false must reject a candidate with using_template=true")
	}
	// Omitted using_template → skip the check, remain a match candidate.
	if !r.applicationMatchesPlanned(context.Background(), candidateTemplate, plannedFalse, false) {
		t.Error("omitted using_template must not reject the candidate")
	}
}
