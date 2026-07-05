package main

import (
	"flag"
	"strings"

	"github.com/K0H205/protolint-plugin-no-shared-repeated/rules"
	"github.com/yoheimuta/protolint/plugin"
)

var allowedTypes = flag.String(
	"allowed_types",
	"",
	"comma-separated list of type names allowed to be shared as repeated fields (e.g. google.protobuf.Timestamp,example.Money)",
)

func main() {
	flag.Parse()

	plugin.RegisterCustomRules(
		rules.NewNoSharedRepeatedMessageRule(splitComma(*allowedTypes)),
	)
}

func splitComma(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}
