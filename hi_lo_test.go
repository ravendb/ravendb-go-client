package ravendb

import (
	"fmt"
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

func TestHiLoCanNotGoDown(t *testing.T) {
	fmt.Printf("TestHiLoCanNotGoDown started\n")
	if useProxy() {
		proxy.ChangeLogFile("trace_hilo_go.txt")
	}
	store, err := getDocumentStore()
	assert.Nil(t, err)
	if store == nil {
		return
	}
}
