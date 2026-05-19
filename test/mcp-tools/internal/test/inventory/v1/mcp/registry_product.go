package mcp

import (
	"github.com/google/jsonschema-go/jsonschema"
	inventoryv1connect "github.com/labset/clarity-protobuf-tools/test/mcp-tools/schema/gen/test/inventory/v1/inventoryv1connect"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type productTools struct {
	handler inventoryv1connect.ProductServiceHandler
}

func RegisterProductTools(server *mcpsdk.Server, handler inventoryv1connect.ProductServiceHandler) {
	t := &productTools{handler: handler}
	server.AddTool(&mcpsdk.Tool{Name: "create_product", Description: "Create Product", InputSchema: &jsonschema.Schema{Type: "object"}}, t.createProduct)
	server.AddTool(&mcpsdk.Tool{Name: "get_product", Description: "Get Product", InputSchema: &jsonschema.Schema{Type: "object"}}, t.getProduct)
	server.AddTool(&mcpsdk.Tool{Name: "list_products", Description: "List Product", InputSchema: &jsonschema.Schema{Type: "object"}}, t.listProducts)
	server.AddTool(&mcpsdk.Tool{Name: "update_product", Description: "Update Product", InputSchema: &jsonschema.Schema{Type: "object"}}, t.updateProduct)
	server.AddTool(&mcpsdk.Tool{Name: "delete_product", Description: "Delete Product", InputSchema: &jsonschema.Schema{Type: "object"}}, t.deleteProduct)
}
