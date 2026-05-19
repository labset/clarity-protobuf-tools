package codegen

import (
	"google.golang.org/protobuf/compiler/protogen"
)

type mcpToolsGenerator struct {
	connectCrud *connectCrudGenerator
}

func (g *mcpToolsGenerator) Generate(plugin *protogen.Plugin) error {
	if err := g.connectCrud.Generate(plugin); err != nil {
		return err
	}

	return nil
}
