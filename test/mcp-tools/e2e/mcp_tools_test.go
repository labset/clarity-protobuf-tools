package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterProductTools(t *testing.T) {
	server := newTestMcpServer()
	assert.NotNil(t, server)
}
