package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

type HiLoDoc struct {
	Max int `json:"Max"`
}

type Product struct {
	ProductName string
}

func TestHiloCanNotGoDown(t *testing.T) {
	if useProxy() {
		proxy.ChangeLogFile("trace_hilo_go.txt")
	}
	store, err := getDocumentStore()
	if store == nil {
		return
	}
	assert.Nil(t, err)
}
