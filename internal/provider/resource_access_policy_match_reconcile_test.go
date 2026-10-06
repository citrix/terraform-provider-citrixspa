package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func arNamed(name, id string, active types.Bool, vals types.List) AccessRuleResourceModel {
	return AccessRuleResourceModel{
		ID:     idOrNull(id),
		Name:   nameOrNull(name),
		Active: active,
		Rules: []RuleResourceModel{{
			Type:     types.StringValue("TYPE_USERGROUP"),
			Operator: types.StringValue("OPERATOR_IN"),
			Values:   vals,
		}},
	}
}

func idOrNull(id string) types.String {
	if id == "" {
		return types.StringNull()
	}
	return types.StringValue(id)
}
func nameOrNull(n string) types.String {
	if n == "" {
		return types.StringNull()
	}
	return types.StringValue(n)
}

// Reorder + ordinary content edit of two uniquely-named rules must not swap their
// backend ids via the positional fallback.
func TestMatchAccessRules_reorderPlusEditNamed_noSwap(t *testing.T) {
	ctx := context.Background()
	tru := types.BoolValue(true)

	state := []AccessRuleResourceModel{
		arNamed("a", "idA", tru, mustList(t, []string{"v1"})),
		arNamed("b", "idB", tru, mustList(t, []string{"v2"})),
	}
	// Reordered AND values edited (concrete), names unchanged & unique.
	plan := []AccessRuleResourceModel{
		arNamed("b", "", tru, mustList(t, []string{"v2-edited"})),
		arNamed("a", "", tru, mustList(t, []string{"v1-edited"})),
	}

	m := matchAccessRulesToPrior(ctx, plan, state, nil)
	g0, g1 := idOf(m[0]), idOf(m[1])
	if g0 != "idB" || g1 != "idA" {
		t.Fatalf("id swap on reorder+edit: plan[0]=b->%s (want idB), plan[1]=a->%s (want idA)", g0, g1)
	}
}

// An unknown top-level signature field (active) on UNNAMED rules makes the
// signature miss; such rules must be excluded from the positional fallback (left
// unmatched) rather than positionally swapped.
func TestMatchAccessRules_unknownActiveUnnamed_notPositional(t *testing.T) {
	ctx := context.Background()
	tru := types.BoolValue(true)
	unk := types.BoolUnknown()

	state := []AccessRuleResourceModel{
		arNamed("", "idX", tru, mustList(t, []string{"v1"})),
		arNamed("", "idY", tru, mustList(t, []string{"v2"})),
	}
	// Reordered, unnamed, active unknown (unresolved output) -> nothing safe to match.
	plan := []AccessRuleResourceModel{
		arNamed("", "", unk, mustList(t, []string{"v2"})),
		arNamed("", "", unk, mustList(t, []string{"v1"})),
	}

	m := matchAccessRulesToPrior(ctx, plan, state, nil)
	if m[0] != nil || m[1] != nil {
		t.Fatalf("unknown-active unnamed rules must stay unmatched, got plan[0]->%s plan[1]->%s", idOf(m[0]), idOf(m[1]))
	}
}

func idOf(r *AccessRuleResourceModel) string {
	if r == nil {
		return "<nil>"
	}
	return r.ID.ValueString()
}

// Two rules identical in every other field but differing only by their
// user-configured description, with names omitted and reordered, must keep their
// ids. description must participate in the content signature so this valid
// reorder stays identity-stable.
func TestMatchAccessRules_descriptionDistinguishesReorder_noSwap(t *testing.T) {
	ctx := context.Background()
	tru := types.BoolValue(true)

	mk := func(id, desc string) AccessRuleResourceModel {
		r := arNamed("", id, tru, mustList(t, []string{"v-same"}))
		r.Description = types.StringValue(desc)
		return r
	}
	state := []AccessRuleResourceModel{mk("idA", "a"), mk("idB", "b")}
	// Reordered; names omitted; only description distinguishes the two rules.
	plan := []AccessRuleResourceModel{mk("", "b"), mk("", "a")}

	m := matchAccessRulesToPrior(ctx, plan, state, nil)
	g0, g1 := idOf(m[0]), idOf(m[1])
	if g0 != "idB" || g1 != "idA" {
		t.Fatalf("description-only reorder swapped ids: plan[0]=b->%s (want idB), plan[1]=a->%s (want idA)", g0, g1)
	}
}

// Reconciliation must not reassign or duplicate a backend rule id when a new
// rule is inserted whose content is identical to an existing rule. The existing
// rule keeps its id and the inserted rule is left unmatched (planned as a fresh
// id) — matchAccessRulesToPrior consumes each prior rule at most once, so a
// still-queued signature cannot be re-claimed by a later duplicate-content rule.
func TestMatchAccessRules_insertedDuplicateContentNoIDInheritance(t *testing.T) {
	ctx := context.Background()
	tru := types.BoolValue(true)

	state := []AccessRuleResourceModel{
		arNamed("keep", "idE", tru, mustList(t, []string{"v-same"})),
	}
	// Keep the existing rule and insert a new one with identical content.
	plan := []AccessRuleResourceModel{
		arNamed("keep", "", tru, mustList(t, []string{"v-same"})),
		arNamed("inserted", "", tru, mustList(t, []string{"v-same"})),
	}

	m := matchAccessRulesToPrior(ctx, plan, state, nil)
	if idOf(m[0]) != "idE" {
		t.Fatalf("existing rule lost its id: got %s want idE", idOf(m[0]))
	}
	if m[1] != nil {
		t.Fatalf("inserted duplicate-content rule inherited an existing id: %s (want new/<nil>)", idOf(m[1]))
	}
}

// A newly inserted, differently-named rule must not steal the content match an
// omitted-name existing rule needs. State [X(name="old")], plan [Y(name="new"),
// X(name omitted)] with identical content: the omitted-name rule keeps idX and
// the named insert gets a fresh id, rather than Y consuming X's identity via the
// content-only pass.
func TestMatchAccessRules_namedInsertDoesNotStealOmittedExisting(t *testing.T) {
	ctx := context.Background()
	tru := types.BoolValue(true)

	state := []AccessRuleResourceModel{
		arNamed("old", "idX", tru, mustList(t, []string{"v-same"})),
	}
	plan := []AccessRuleResourceModel{
		arNamed("new", "", tru, mustList(t, []string{"v-same"})),
		arNamed("", "", tru, mustList(t, []string{"v-same"})),
	}

	m := matchAccessRulesToPrior(ctx, plan, state, nil)
	if m[0] != nil {
		t.Fatalf("inserted named rule stole a prior id: %s (want new/<nil>)", idOf(m[0]))
	}
	if idOf(m[1]) != "idX" {
		t.Fatalf("omitted-name existing rule lost its id: got %s want idX", idOf(m[1]))
	}
}

// consumePriorMeta: an explicit plan value must remove the equal prior entry (not
// blindly FIFO the first), so an omitted same-key sibling still adopts the other
// prior value instead of both collapsing onto the explicit one.
func TestConsumePriorMeta_explicitConsumesEqualNotFIFO(t *testing.T) {
	a := mustMap(t, map[string]string{"k": "A"})
	b := mustMap(t, map[string]string{"k": "B"})
	q := []types.Map{a, b}

	// Explicit B removes B, leaving [A].
	adopt, rest := consumePriorMeta(q, b)
	if adopt != nil {
		t.Fatalf("explicit value must not adopt prior metadata, got %v", adopt)
	}
	if len(rest) != 1 || !rest[0].Equal(a) {
		t.Fatalf("explicit consume should leave [A], got %v", rest)
	}
	// Omitted sibling then adopts the remaining A.
	adopt2, rest2 := consumePriorMeta(rest, types.MapNull(types.StringType))
	if adopt2 == nil || !adopt2.Equal(a) {
		t.Fatalf("omitted sibling should adopt A, got %v", adopt2)
	}
	if len(rest2) != 0 {
		t.Fatalf("queue should be empty after both consumed, got %v", rest2)
	}
}

// consumePriorMeta: when the explicit plan value CHANGED (no prior is equal to
// it), it must consume its positional slot (the queue head) so a following
// omitted same-key sibling adopts the next prior value, not the changed
// sibling's old metadata. Prior [A,B]; explicit C then omitted -> omitted gets B.
func TestConsumePriorMeta_explicitChangedConsumesHead(t *testing.T) {
	a := mustMap(t, map[string]string{"k": "A"})
	b := mustMap(t, map[string]string{"k": "B"})
	c := mustMap(t, map[string]string{"k": "C"})
	q := []types.Map{a, b}

	// Explicit changed value C matches no prior; it consumes the head A, leaving [B].
	adopt, rest := consumePriorMeta(q, c)
	if adopt != nil {
		t.Fatalf("explicit value must not adopt prior metadata, got %v", adopt)
	}
	if len(rest) != 1 || !rest[0].Equal(b) {
		t.Fatalf("changed explicit should consume head A leaving [B], got %v", rest)
	}
	// Omitted sibling then adopts B (not A, the changed sibling's old value).
	adopt2, rest2 := consumePriorMeta(rest, types.MapNull(types.StringType))
	if adopt2 == nil || !adopt2.Equal(b) {
		t.Fatalf("omitted sibling should adopt B, got %v", adopt2)
	}
	if len(rest2) != 0 {
		t.Fatalf("queue should be empty after both consumed, got %v", rest2)
	}
}

// An explicitly-configured rule id must be authoritative: when two rules are
// reordered AND content-edited together with their names omitted, the content
// and name passes miss and the positional fallback would pair each plan rule
// with the wrong prior at its new index. Passing the explicit ids (trustedIDs)
// makes Pass 0 consume the right prior first, so each rule keeps its own
// identity. The nil-trustedIDs call documents the mispairing the id pass fixes.
func TestMatchAccessRules_explicitIDAuthoritativeOnReorderEdit(t *testing.T) {
	ctx := context.Background()
	tru := types.BoolValue(true)

	state := []AccessRuleResourceModel{
		arNamed("a", "idA", tru, mustList(t, []string{"v1"})),
		arNamed("b", "idB", tru, mustList(t, []string{"v2"})),
	}
	// Reordered (b first), content-edited (values changed), names omitted, but
	// each rule still carries its explicit backend id.
	plan := []AccessRuleResourceModel{
		arNamed("", "idB", tru, mustList(t, []string{"v2-edited"})),
		arNamed("", "idA", tru, mustList(t, []string{"v1-edited"})),
	}
	trusted := []string{"idB", "idA"}

	m := matchAccessRulesToPrior(ctx, plan, state, trusted)
	if idOf(m[0]) != "idB" || idOf(m[1]) != "idA" {
		t.Fatalf("explicit id not authoritative: got [%s %s] want [idB idA]", idOf(m[0]), idOf(m[1]))
	}

	// Without the id pass the positional fallback mispairs (each plan rule takes
	// the prior at its new index), which is exactly the drift the id pass avoids.
	bad := matchAccessRulesToPrior(ctx, plan, state, nil)
	if idOf(bad[0]) != "idA" || idOf(bad[1]) != "idB" {
		t.Fatalf("expected positional mispairing without id pass, got [%s %s]", idOf(bad[0]), idOf(bad[1]))
	}
}

// A plan rule with an explicit id that matches no prior is a genuinely new rule
// with a chosen id; it must not fall through to the content pass and steal an
// existing rule's identity, even when their content is identical.
func TestMatchAccessRules_explicitNewIDNotContentStolen(t *testing.T) {
	ctx := context.Background()
	tru := types.BoolValue(true)

	state := []AccessRuleResourceModel{
		arNamed("", "idX", tru, mustList(t, []string{"v-same"})),
	}
	plan := []AccessRuleResourceModel{
		arNamed("", "idNEW", tru, mustList(t, []string{"v-same"})),
	}

	m := matchAccessRulesToPrior(ctx, plan, state, []string{"idNEW"})
	if m[0] != nil {
		t.Fatalf("explicit new-id rule stole an existing prior by content: %s (want <nil>)", idOf(m[0]))
	}
}

// Same guarantee when the existing rule is matched by content signature alone
// (name omitted): the duplicate inserted rule still gets no prior match.
func TestMatchAccessRules_insertedDuplicateContentUnnamed_noInheritance(t *testing.T) {
	ctx := context.Background()
	tru := types.BoolValue(true)

	state := []AccessRuleResourceModel{
		arNamed("", "idE", tru, mustList(t, []string{"v-same"})),
	}
	plan := []AccessRuleResourceModel{
		arNamed("", "", tru, mustList(t, []string{"v-same"})),
		arNamed("", "", tru, mustList(t, []string{"v-same"})),
	}

	m := matchAccessRulesToPrior(ctx, plan, state, nil)
	if idOf(m[0]) != "idE" {
		t.Fatalf("existing rule lost its id: got %s want idE", idOf(m[0]))
	}
	if m[1] != nil {
		t.Fatalf("inserted duplicate-content rule inherited an existing id: %s (want new/<nil>)", idOf(m[1]))
	}
}

// A prior consumed by the explicit-id pass must not be re-matched by the
// name+content pass. State [n/idR]; plan [n id-pinned to idR, n new (same
// name+content)]. The id-pinned rule takes idR and the new rule must stay
// unmatched, not receive idR a second time (which would PUT one id twice).
func TestMatchAccessRules_pass1DoesNotReuseIDConsumedPrior(t *testing.T) {
	ctx := context.Background()
	tru := types.BoolValue(true)

	state := []AccessRuleResourceModel{
		arNamed("n", "idR", tru, mustList(t, []string{"v1"})),
	}
	plan := []AccessRuleResourceModel{
		arNamed("n", "", tru, mustList(t, []string{"v1"})),
		arNamed("n", "", tru, mustList(t, []string{"v1"})),
	}

	m := matchAccessRulesToPrior(ctx, plan, state, []string{"idR", ""})
	if idOf(m[0]) != "idR" {
		t.Fatalf("id-pinned rule lost idR: got %s", idOf(m[0]))
	}
	if m[1] != nil {
		t.Fatalf("new same-name/content rule reused already-consumed idR: %s (want <nil>)", idOf(m[1]))
	}
}

// Renaming AND reordering two rules in the same plan must not swap their ids.
// State [A/idA(X), B/idB(Y)]; plan [B2(Y), A2(X)] (both renamed, reordered). The
// name passes miss, so content recovery (Pass 3b) must keep each rule's id with
// its content rather than the positional fallback swapping them.
func TestMatchAccessRules_renamePlusReorder_noSwap(t *testing.T) {
	ctx := context.Background()
	tru := types.BoolValue(true)

	state := []AccessRuleResourceModel{
		arNamed("A", "idA", tru, mustList(t, []string{"X"})),
		arNamed("B", "idB", tru, mustList(t, []string{"Y"})),
	}
	plan := []AccessRuleResourceModel{
		arNamed("B2", "", tru, mustList(t, []string{"Y"})),
		arNamed("A2", "", tru, mustList(t, []string{"X"})),
	}

	m := matchAccessRulesToPrior(ctx, plan, state, nil)
	if idOf(m[0]) != "idB" || idOf(m[1]) != "idA" {
		t.Fatalf("rename+reorder swapped ids: plan[0](Y)->%s (want idB), plan[1](X)->%s (want idA)", idOf(m[0]), idOf(m[1]))
	}
}

// Pass 3b must not override a surviving unique-name match. State [A/idA(X),
// B/idB(Y)]; plan single rule named A with content edited to Y. The name A still
// exists in state, so the rule must keep idA (Pass 3), not adopt idB by content.
func TestMatchAccessRules_survivingNameBeatsContentRecovery(t *testing.T) {
	ctx := context.Background()
	tru := types.BoolValue(true)

	state := []AccessRuleResourceModel{
		arNamed("A", "idA", tru, mustList(t, []string{"X"})),
		arNamed("B", "idB", tru, mustList(t, []string{"Y"})),
	}
	plan := []AccessRuleResourceModel{
		arNamed("A", "", tru, mustList(t, []string{"Y"})),
	}

	m := matchAccessRulesToPrior(ctx, plan, state, nil)
	if idOf(m[0]) != "idA" {
		t.Fatalf("surviving name lost to content recovery: got %s want idA", idOf(m[0]))
	}
}

// arMeta builds an unnamed rule with a single sub-rule carrying explicit metadata.
func arMeta(id string, vals types.List, meta types.Map) AccessRuleResourceModel {
	return AccessRuleResourceModel{
		ID:     idOrNull(id),
		Name:   types.StringNull(),
		Active: types.BoolValue(true),
		Rules: []RuleResourceModel{{
			Type:     types.StringValue("TYPE_USERGROUP"),
			Operator: types.StringValue("OPERATOR_IN"),
			Values:   vals,
			Metadata: meta,
		}},
	}
}

// Two unnamed rules identical except for distinct, explicitly-configured metadata
// must not swap their ids when reordered: the metadata-inclusive pass (0b)
// distinguishes them before the metadata-independent FIFO pass runs.
func TestMatchAccessRules_explicitMetadataDistinguishesReorder(t *testing.T) {
	ctx := context.Background()
	M0 := mustMap(t, map[string]string{"d": "0"})
	M1 := mustMap(t, map[string]string{"d": "1"})

	state := []AccessRuleResourceModel{
		arMeta("id0", mustList(t, []string{"v"}), M0),
		arMeta("id1", mustList(t, []string{"v"}), M1),
	}
	// Reordered: plan[0] carries M1 (was id1), plan[1] carries M0 (was id0).
	plan := []AccessRuleResourceModel{
		arMeta("", mustList(t, []string{"v"}), M1),
		arMeta("", mustList(t, []string{"v"}), M0),
	}

	m := matchAccessRulesToPrior(ctx, plan, state, nil)
	if idOf(m[0]) != "id1" || idOf(m[1]) != "id0" {
		t.Fatalf("explicit metadata reorder swapped ids: plan[0](M1)->%s (want id1), plan[1](M0)->%s (want id0)", idOf(m[0]), idOf(m[1]))
	}
}

// A rule with omitted metadata must still match its prior by content when the
// prior has known metadata: Pass 0b must be skipped for omitted-metadata rules so
// the metadata-independent passes keep matching them (id retained on reorder).
func TestMatchAccessRules_omittedMetadataStillMatchesByContent(t *testing.T) {
	ctx := context.Background()
	tru := types.BoolValue(true)

	state := []AccessRuleResourceModel{
		arMeta("idM", mustList(t, []string{"v"}), mustMap(t, map[string]string{"d": "keep"})),
	}
	// Same values, name omitted, metadata omitted (null) — reconciled at plan.
	plan := []AccessRuleResourceModel{
		arNamed("", "", tru, mustList(t, []string{"v"})),
	}

	m := matchAccessRulesToPrior(ctx, plan, state, nil)
	if idOf(m[0]) != "idM" {
		t.Fatalf("omitted-metadata rule lost its id: got %s want idM", idOf(m[0]))
	}
}
