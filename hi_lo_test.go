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

func (d *HiLoDoc) getMax() int {
	return d.Max
}

func (d *HiLoDoc) setMax(max int) {
	d.Max = max
}

type Product struct {
	ProductName string `json:"ProductName"`
}

func (p *Product) getProductName() String {
	return p.ProductName
}

func (p *Product) setProductName(productName String) {
	p.ProductName = productName
}

func TestCapacityShouldDouble(t *testing.T) {
	fmt.Printf("TestHiLoCanNotGoDown started\n")
	if useProxy() {
		proxy.ChangeLogFile("trace_hilo_go.txt")
	}
	store, err := getDocumentStore()
	assert.Nil(t, err)
	if store == nil {
		return // if db tests are disabled
	}
	hiLoIdGenerator := NewHiLoIdGenerator("users", store, store.getDatabase(), store.getConventions().getIdentityPartsSeparator())

	session, err := store.OpenSession()
	assert.Nil(t, err)
	assert.NotNil(t, session)
	hiloDoc := &HiLoDoc{}
	hiloDoc.setMax(64)

	session.StoreEntityWithID(hiloDoc, "Raven/Hilo/users")
	err = session.SaveChanges()
	assert.Nil(t, err)

	for i := 0; i < 32; i++ {
		hiLoIdGenerator.GenerateDocumentID(&User{})
	}

}

/*
func TestHiLoCanNotGoDown(t *testing.T) {
	fmt.Printf("TestHiLoCanNotGoDown started\n")
	if useProxy() {
		proxy.ChangeLogFile("trace_hilo_go.txt")
	}
	store, err := getDocumentStore()
	assert.Nil(t, err)
	if store == nil {
		return // if db tests are disabled
	}

	session, err := store.OpenSession()
	assert.Nil(t, err)
	assert.NotNil(t, session)
	hiloDoc := &HiLoDoc{}
	hiloDoc.setMax(32)

	session.StoreEntityWithID(hiloDoc, "Raven/Hilo/users")
	err = session.SaveChanges()
	assert.Nil(t, err)
}
*/
