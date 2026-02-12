package main

import (
	"github.com/K0H205/protolint-plugin-no-shared-repeated/rules"
	"github.com/yoheimuta/protolint/plugin"
)

func main() {
	plugin.RegisterCustomRules(
		rules.NewNoSharedRepeatedMessageRule(),
	)
}
