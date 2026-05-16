package main

import (
	"buf.build/go/bufplugin/check"
	"github.com/labset/clarity-protobuf-tools/internal/rules"
)

func main() {
	check.Main(&check.Spec{
		Rules: rules.All,
	})
}
