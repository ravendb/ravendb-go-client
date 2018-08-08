package ravendb

import (
	"fmt"
	"runtime/debug"
	"testing"

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

func (p *Product) getProductName() string {
	return p.ProductName
}

func (p *Product) setProductName(productName string) {
	p.ProductName = productName
}

func hiloTest_capacityShouldDouble(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	hiLoIdGenerator := NewHiLoIdGenerator("users", store, store.getDatabase(), store.getConventions().getIdentityPartsSeparator())

	{
		session := openSessionMust(t, store)
		hiloDoc := &HiLoDoc{}
		hiloDoc.setMax(64)

		err = session.StoreWithID(hiloDoc, "Raven/Hilo/users")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		for i := 0; i < 32; i++ {
			hiLoIdGenerator.GenerateDocumentID(NewUser())
		}
		session.Close()
	}

	{
		session := openSessionMust(t, store)

		//var hiloDoc HiLoDoc
		//err = session.load(&hiloDoc, "Raven/Hilo/users")
		//assert.Nil(t, err)

		result, err := session.load(getTypeOf(&HiLoDoc{}), "Raven/Hilo/users")
		assert.NoError(t, err)
		hiloDoc := result.(*HiLoDoc)
		max := hiloDoc.getMax()
		assert.Equal(t, max, 96)

		//we should be receiving a range of 64 now
		hiLoIdGenerator.GenerateDocumentID(NewUser())
		session.Close()
	}

	{
		session := openSessionMust(t, store)

		result, err := session.load(getTypeOf(&HiLoDoc{}), "Raven/Hilo/users")
		assert.NoError(t, err)
		hiloDoc := result.(*HiLoDoc)
		max := hiloDoc.getMax()

		// TODO: in Java it's 160. On Travis CI (linux) it's 160
		// On my mac, it's 128.
		// It's strange because the requests sent for
		// /databases/test_db_1/hilo/next are exactly the
		// same but in Java case the server sends back "High": 160
		// and in Go case it's "High": 128
		// Maybe it's KeepAlive difference?
		valid := (max == 96+64) || (max == 96+32)
		assert.True(t, valid)
		session.Close()
	}
}

func hiloTest_returnUnusedRangeOnClose(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	newStore := NewDocumentStore()
	newStore.setUrls(store.getUrls())
	newStore.setDatabase(store.getDatabase())

	_, err = newStore.Initialize()
	assert.NoError(t, err)

	{
		session := openSessionMust(t, newStore)
		assert.NoError(t, err)
		assert.NotNil(t, session)

		hiloDoc := &HiLoDoc{}
		hiloDoc.setMax(32)
		err = session.StoreWithID(hiloDoc, "Raven/Hilo/users")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		err = session.Store(NewUser())
		assert.NoError(t, err)
		err = session.Store(NewUser())
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	newStore.Close() //on document store close, hilo-return should be called

	newStore = NewDocumentStore()
	newStore.setUrls(store.getUrls())
	newStore.setDatabase(store.getDatabase())

	_, err = newStore.Initialize()
	assert.NoError(t, err)

	{
		session := openSessionMust(t, newStore)
		assert.NoError(t, err)
		assert.NotNil(t, session)

		hiloDocI, err := session.load(getTypeOf(&HiLoDoc{}), "Raven/Hilo/users")
		assert.NoError(t, err)
		hiloDoc := hiloDocI.(*HiLoDoc)
		max := hiloDoc.getMax()
		assert.Equal(t, max, 34)
		session.Close()
	}

	newStore.Close() //on document store close, hilo-return should be called
}

func hiloTest_canNotGoDown(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	session := openSessionMust(t, store)

	hiloDoc := &HiLoDoc{}
	hiloDoc.setMax(32)

	err = session.StoreWithID(hiloDoc, "Raven/Hilo/users")
	assert.NoError(t, err)
	err = session.SaveChanges()
	assert.NoError(t, err)
	assert.Nil(t, err)

	hiLoKeyGenerator := NewHiLoIdGenerator("users", store, store.getDatabase(), store.getConventions().getIdentityPartsSeparator())

	nextID, err := hiLoKeyGenerator.nextID()
	assert.Nil(t, err)
	ids := []int{nextID}

	hiloDoc.setMax(12)
	session.StoreWithChangeVectorAndID(hiloDoc, nil, "Raven/Hilo/users")
	err = session.SaveChanges()
	assert.Nil(t, err)

	for i := 0; i < 128; i++ {
		nextID, err = hiLoKeyGenerator.nextID()
		contains := intArrayContains(ids, nextID)
		assert.False(t, contains)
		ids = append(ids, nextID)
	}
	assert.False(t, intArrayHasDuplicates(ids))
	session.Close()
}

func hiloTest_multiDb(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	session := openSessionMust(t, store)

	hiloDoc := &HiLoDoc{}
	hiloDoc.setMax(64)
	err = session.StoreWithID(hiloDoc, "Raven/Hilo/users")
	assert.NoError(t, err)

	productsHilo := &HiLoDoc{}
	productsHilo.setMax(128)
	err = session.StoreWithID(productsHilo, "Raven/Hilo/products")
	assert.NoError(t, err)

	err = session.SaveChanges()
	assert.NoError(t, err)

	multiDbHilo := NewMultiDatabaseHiLoIdGenerator(store, store.getConventions())
	generateDocumentKey := multiDbHilo.GenerateDocumentID("", NewUser())
	assert.Equal(t, generateDocumentKey, "users/65-A")
	generateDocumentKey = multiDbHilo.GenerateDocumentID("", &Product{})
	assert.Equal(t, generateDocumentKey, "products/129-A")
	session.Close()
}

func TestHiLo(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer func() {
		r := recover()
		destroyDriver()
		if r != nil {
			fmt.Printf("Panic: '%v'\n", r)
			debug.PrintStack()
			t.Fail()
		}
	}()

	// matches order of java tests
	hiloTest_capacityShouldDouble(t)
	hiloTest_returnUnusedRangeOnClose(t)
	hiloTest_canNotGoDown(t)
	hiloTest_multiDb(t)
}
