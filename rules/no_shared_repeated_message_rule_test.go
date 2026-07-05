package rules_test

import (
	"os"
	"strings"
	"testing"

	"github.com/K0H205/protolint-plugin-no-shared-repeated/rules"
	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/protolint/linter/report"
)

// applyRule parses the given testdata file and applies the rule to it.
func applyRule(t *testing.T, r rules.NoSharedRepeatedMessageRule, path string) []report.Failure {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open %s: %v", path, err)
	}
	defer f.Close()

	proto, err := protoparser.Parse(f)
	if err != nil {
		t.Fatalf("failed to parse %s: %v", path, err)
	}

	failures, err := r.Apply(proto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return failures
}

func expectNoFailures(t *testing.T, failures []report.Failure, context string) {
	t.Helper()
	if len(failures) != 0 {
		t.Errorf("expected no failures for %s, got %d:", context, len(failures))
		for _, f := range failures {
			t.Errorf("  %s", f)
		}
	}
}

func TestNoSharedRepeatedMessageRule_OK(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule(nil)
	failures := applyRule(t, r, "../testdata/ok.proto")
	expectNoFailures(t, failures, "ok.proto")
}

func TestNoSharedRepeatedMessageRule_NG(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule(nil)
	failures := applyRule(t, r, "../testdata/ng.proto")

	if len(failures) != 2 {
		t.Fatalf("expected 2 failures for ng.proto, got %d", len(failures))
	}
	for _, f := range failures {
		msg := f.Message()
		if msg == "" {
			t.Error("failure message should not be empty")
		}
		t.Logf("failure: %s", f)
	}
}

func TestNoSharedRepeatedMessageRule_SameMessageMultipleFields(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule(nil)
	failures := applyRule(t, r, "../testdata/same_message_multiple_fields.proto")
	// Same type used as repeated twice in the SAME message should not trigger.
	expectNoFailures(t, failures, "same_message_multiple_fields.proto")
}

func TestNoSharedRepeatedMessageRule_NestedMessages(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule(nil)
	failures := applyRule(t, r, "../testdata/nested.proto")

	// ItemInfo used as repeated in Outer and Outer.Inner → 2 different messages → 2 failures
	if len(failures) != 2 {
		t.Fatalf("expected 2 failures for nested messages, got %d", len(failures))
	}
	for _, f := range failures {
		if !strings.Contains(f.Message(), `"Outer.Inner"`) {
			t.Errorf("failure message should name the nested parent by full path, got: %s", f.Message())
		}
	}
}

func TestNoSharedRepeatedMessageRule_QualifiedNames(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule(nil)
	failures := applyRule(t, r, "../testdata/qualified.proto")

	// "example.ItemInfo" and "ItemInfo" resolve to the same type → 2 failures.
	if len(failures) != 2 {
		t.Fatalf("expected 2 failures for qualified.proto, got %d", len(failures))
	}
	for _, f := range failures {
		t.Logf("failure: %s", f)
	}
}

func TestNoSharedRepeatedMessageRule_SameNameDifferentScope(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule(nil)
	failures := applyRule(t, r, "../testdata/same_name_different_scope.proto")
	// A.Item and B.Item are distinct types despite the same short name.
	expectNoFailures(t, failures, "same_name_different_scope.proto")
}

func TestNoSharedRepeatedMessageRule_EnumNotReported(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule(nil)
	failures := applyRule(t, r, "../testdata/enum.proto")
	expectNoFailures(t, failures, "enum.proto")
}

func TestNoSharedRepeatedMessageRule_ExternalTypeReported(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule(nil)
	failures := applyRule(t, r, "../testdata/external.proto")

	if len(failures) != 2 {
		t.Fatalf("expected 2 failures for external.proto without allowlist, got %d", len(failures))
	}
}

func TestNoSharedRepeatedMessageRule_AllowedTypes(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule([]string{"google.protobuf.Timestamp"})
	failures := applyRule(t, r, "../testdata/external.proto")
	expectNoFailures(t, failures, "external.proto with allowlist")

	r = rules.NewNoSharedRepeatedMessageRule([]string{"ItemInfo"})
	failures = applyRule(t, r, "../testdata/ng.proto")
	expectNoFailures(t, failures, "ng.proto with ItemInfo allowlisted")
}

func TestNoSharedRepeatedMessageRule_DeterministicOrder(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule(nil)

	// Two violating types in one file; failures must always come out in
	// file-position order regardless of map iteration order.
	for i := 0; i < 10; i++ {
		failures := applyRule(t, r, "../testdata/multiple_types.proto")
		if len(failures) != 4 {
			t.Fatalf("expected 4 failures for multiple_types.proto, got %d", len(failures))
		}
		for j := 1; j < len(failures); j++ {
			prev, cur := failures[j-1].Pos(), failures[j].Pos()
			if prev.Offset > cur.Offset {
				t.Fatalf("failures are not in file-position order: %v comes after %v", prev, cur)
			}
		}
	}
}

func TestNoSharedRepeatedMessageRule_ID(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule(nil)
	if got := r.ID(); got != "NO_SHARED_REPEATED_MESSAGE" {
		t.Errorf("ID() = %q, want %q", got, "NO_SHARED_REPEATED_MESSAGE")
	}
}

func TestNoSharedRepeatedMessageRule_Purpose(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule(nil)
	if got := r.Purpose(); got == "" {
		t.Error("Purpose() should not be empty")
	}
}

func TestNoSharedRepeatedMessageRule_Severity(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule(nil)
	if got := r.Severity(); got != "warning" {
		t.Errorf("Severity() = %q, want %q", got, "warning")
	}
}
