package xpath

import (
	"math"
	"testing"
)

// Synthetic ledger-style document for numeric operand context isolation.
// Two Credit siblings (Amounts 20 and 5) so one compiled expression can be
// evaluated across different record contexts.
func createCreditDoc() *TNode {
	// <Root>
	//   <Credit><Entry><Amount>20</Amount></Entry></Credit>
	//   <Credit><Entry><Amount>5</Amount></Entry></Credit>
	// </Root>
	doc := createNode("", RootNode)
	root := doc.createChildNode("Root", ElementNode)
	for _, amount := range []string{"20", "5"} {
		ev := root.createChildNode("Credit", ElementNode)
		entry := ev.createChildNode("Entry", ElementNode)
		amt := entry.createChildNode("Amount", ElementNode)
		amt.createChildNode(amount, TextNode)
	}
	return doc
}

func selectCredit(doc *TNode) *TNode {
	return selectNode(doc, "//Credit")
}

func selectCredits(doc *TNode) []*TNode {
	return selectNodes(doc, "//Credit")
}

func evalNumberAt(t *testing.T, context *TNode, expr string) float64 {
	t.Helper()
	exp, err := Compile(expr)
	if err != nil {
		t.Fatalf("Compile(%q): %v", expr, err)
	}
	nav := createNavigator(context)
	v := exp.Evaluate(nav)
	f, ok := v.(float64)
	if !ok {
		t.Fatalf("Evaluate(%q) type %T want float64 (value=%v)", expr, v, v)
	}
	return f
}

func assertNear(t *testing.T, got, want float64, expr string) {
	t.Helper()
	if math.IsNaN(got) || math.IsNaN(want) {
		if !(math.IsNaN(got) && math.IsNaN(want)) {
			t.Fatalf("%s: got %v want %v", expr, got, want)
		}
		return
	}
	if got != want {
		t.Fatalf("%s: got %v want %v", expr, got, want)
	}
}

// TestNumericOperatorContextIsolation covers the predicated-sum × context-sensitive
// factor class: predicates move the shared navigator; the right operand must still
// observe the operator's original context.
func TestNumericOperatorContextIsolation(t *testing.T) {
	doc := createCreditDoc()
	ctx := selectCredit(doc)
	if ctx == nil {
		t.Fatal("Credit context not found")
	}

	const (
		predSumLit  = `sum(.//Entry[not(@excluded='true')]/Amount) * -1`
		plainSumCS  = `sum(.//Amount) * (1 - 2*count(self::Credit))`
		predSumCS   = `sum(.//Entry[not(@excluded='true')]/Amount) * (1 - 2*count(self::Credit))`
		factorFirst = `(1 - 2*count(self::Credit)) * sum(.//Entry[not(@excluded='true')]/Amount)`
		// Magnitude > 1 non-identity factor
		predSumMag = `sum(.//Entry[not(@excluded='true')]/Amount) * (3 * count(self::Credit))`
		// Div and minus neighbors sharing numericQuery
		predSumDiv = `sum(.//Entry[not(@excluded='true')]/Amount) div (count(self::Credit) + 1)`
		predSumSub = `sum(.//Entry[not(@excluded='true')]/Amount) - count(self::Credit)`
	)

	// Controls and the defect row (must all be correct after the fix).
	cases := []struct {
		expr string
		want float64
	}{
		{predSumLit, -20},
		{plainSumCS, -20},
		{predSumCS, -20},
		{factorFirst, -20},
		{predSumMag, 60},
		{predSumDiv, 10},
		{predSumSub, 19},
	}
	for _, tc := range cases {
		got := evalNumberAt(t, ctx, tc.expr)
		assertNear(t, got, tc.want, tc.expr)
	}

	// Original context navigator must be unchanged after full evaluation.
	nav := createNavigator(ctx)
	beforeName := nav.LocalName()
	exp := MustCompile(predSumCS)
	_ = exp.Evaluate(nav)
	if nav.LocalName() != beforeName {
		t.Fatalf("context navigator moved after Evaluate: got %q want %q", nav.LocalName(), beforeName)
	}

	// Same compiled expression: same-context repeat (fresh navigator each time).
	exp2 := MustCompile(predSumCS)
	gSame1 := exp2.Evaluate(createNavigator(ctx)).(float64)
	gSame2 := exp2.Evaluate(createNavigator(ctx)).(float64)
	if gSame1 != -20 || gSame2 != -20 {
		t.Fatalf("same-context repeat Evaluate: g1=%v g2=%v want both -20", gSame1, gSame2)
	}

	// Different record contexts must not contaminate query/navigator state.
	// One compiled expression, contexts A (Amount=20) → B (Amount=5) → A.
	credits := selectCredits(doc)
	if len(credits) != 2 {
		t.Fatalf("expected 2 Credit contexts, got %d", len(credits))
	}
	ctxA, ctxB := credits[0], credits[1]
	expAB := MustCompile(predSumCS)
	vA1 := expAB.Evaluate(createNavigator(ctxA))
	vB := expAB.Evaluate(createNavigator(ctxB))
	vA2 := expAB.Evaluate(createNavigator(ctxA))
	gotA1, okA1 := vA1.(float64)
	gotB, okB := vB.(float64)
	gotA2, okA2 := vA2.(float64)
	if !okA1 || !okB || !okA2 {
		t.Fatalf("cross-context Evaluate types: A1=%T B=%T A2=%T", vA1, vB, vA2)
	}
	if gotA1 != -20 || gotB != -5 || gotA2 != -20 {
		t.Fatalf("cross-context Evaluate contamination: A1=%v B=%v A2=%v want -20, -5, -20", gotA1, gotB, gotA2)
	}

	// Empty node-set: sum empty = 0; context-sensitive factor still non-identity.
	emptyExpr := `sum(.//Entry[@missing='yes']/Amount) * (1 - 2*count(self::Credit))`
	gotEmpty := evalNumberAt(t, ctx, emptyExpr)
	assertNear(t, gotEmpty, 0, emptyExpr) // 0 * -1 = 0
}

// TestNumericOperatorNodeSetCoercion ensures a bare node-set left operand still
// converts correctly when the right operand is context-sensitive (eager asNumber
// path after the isolation fix).
func TestNumericOperatorNodeSetCoercion(t *testing.T) {
	doc := createCreditDoc()
	ctx := selectCredit(doc)
	// First Amount node value is 20; times context-sensitive -1.
	expr := `.//Amount * (1 - 2*count(self::Credit))`
	got := evalNumberAt(t, ctx, expr)
	assertNear(t, got, -20, expr)
}
