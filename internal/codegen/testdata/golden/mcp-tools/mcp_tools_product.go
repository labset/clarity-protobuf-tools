package api

import (
	"github.com/acme/app/internal/acme/inventory/v1/db"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type productTools struct {
	handler *productHandler
}

func RegisterProductTools(server *mcp.Server, deps ProductDeps) {
	t := &productTools{
		handler: &productHandler{
			store: db.New(deps.Pool),
		},
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_product",
		Description: "Create a new product",
	}, t.createProduct)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_product",
		Description: "Get a product by ID",
	}, t.getProduct)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_products",
		Description: "List products",
	}, t.listProducts)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_product",
		Description: "Update a product",
	}, t.updateProduct)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_product",
		Description: "Delete a product",
	}, t.deleteProduct)
}
