package rules_test

import (
	"os"
	"testing"

	"github.com/K0H205/protolint-plugin-no-shared-repeated/rules"
	"github.com/yoheimuta/go-protoparser/v4"
)

func TestNoSharedRepeatedMessageRule_OK(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule()

	f, err := os.Open("../testdata/ok.proto")
	if err != nil {
		t.Fatalf("failed to open ok.proto: %v", err)
	}
	defer f.Close()

	proto, err := protoparser.Parse(f)
	if err != nil {
		t.Fatalf("failed to parse ok.proto: %v", err)
	}

	failures, err := r.Apply(proto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(failures) != 0 {
		t.Errorf("expected no failures for ok.proto, got %d:", len(failures))
		for _, f := range failures {
			t.Errorf("  %s", f)
		}
	}
}

func TestNoSharedRepeatedMessageRule_NG(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule()

	f, err := os.Open("../testdata/ng.proto")
	if err != nil {
		t.Fatalf("failed to open ng.proto: %v", err)
	}
	defer f.Close()

	proto, err := protoparser.Parse(f)
	if err != nil {
		t.Fatalf("failed to parse ng.proto: %v", err)
	}

	failures, err := r.Apply(proto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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
	r := rules.NewNoSharedRepeatedMessageRule()

	f, err := os.Open("../testdata/same_message_multiple_fields.proto")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	proto, err := protoparser.Parse(f)
	if err != nil {
		t.Fatalf("failed to parse proto: %v", err)
	}

	failures, err := r.Apply(proto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Same type used as repeated twice in the SAME message should not trigger.
	if len(failures) != 0 {
		t.Errorf("expected no failures when same type repeated in same message, got %d", len(failures))
		for _, f := range failures {
			t.Errorf("  %s", f)
		}
	}
}

func TestNoSharedRepeatedMessageRule_NestedMessages(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule()

	f, err := os.Open("../testdata/nested.proto")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	proto, err := protoparser.Parse(f)
	if err != nil {
		t.Fatalf("failed to parse proto: %v", err)
	}

	failures, err := r.Apply(proto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// ItemInfo used as repeated in Outer and Outer.Inner → 2 different messages → 2 failures
	if len(failures) != 2 {
		t.Fatalf("expected 2 failures for nested messages, got %d", len(failures))
	}
	for _, f := range failures {
		t.Logf("failure: %s", f)
	}
}

func TestNoSharedRepeatedMessageRule_ID(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule()
	if got := r.ID(); got != "NO_SHARED_REPEATED_MESSAGE" {
		t.Errorf("ID() = %q, want %q", got, "NO_SHARED_REPEATED_MESSAGE")
	}
}

func TestNoSharedRepeatedMessageRule_Purpose(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule()
	if got := r.Purpose(); got == "" {
		t.Error("Purpose() should not be empty")
	}
}

func TestNoSharedRepeatedMessageRule_Severity(t *testing.T) {
	r := rules.NewNoSharedRepeatedMessageRule()
	if got := r.Severity(); got != "warning" {
		t.Errorf("Severity() = %q, want %q", got, "warning")
	}
}
