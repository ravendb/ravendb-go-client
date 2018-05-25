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
	assert.Nil(t, err)
	assert.Nil(t, store)
}
