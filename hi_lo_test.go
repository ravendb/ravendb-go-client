package ravendb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type HiLoDoc struct {
	Max int `json:"Max"`
}

type Product struct {
	ProductName string
}

func TestHiloCanNotGoDown(t *testing.T) {
	store, err := getDocumentStore()
	if store == nil {
		return
	}
	assert.Nil(t, err)
}
