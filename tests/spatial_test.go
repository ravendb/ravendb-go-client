package tests

import (
	"reflect"
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

type MyDocumentItem struct {
	Date      time.Time `json:"date"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
}

type MyDocument struct {
	ID    string
	Items []*MyDocumentItem `json:"items"`
}

type MyProjection struct {
	ID        string
	Date      time.Time `json:"date"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
}

func NewMyIndex() *ravendb.IndexCreationTask {
	res := ravendb.NewIndexCreationTask("MyIndex")
	res.Map = "docs.MyDocuments.SelectMany(doc => doc.items, (doc, item) => new {\n" +
		"    doc = doc,\n" +
		"    item = item\n" +
		"}).Select(this0 => new {\n" +
		"    this0 = this0,\n" +
		"    lat = ((double)(this0.item.latitude ?? 0))\n" +
		"}).Select(this1 => new {\n" +
		"    this1 = this1,\n" +
		"    lng = ((double)(this1.this0.item.longitude ?? 0))\n" +
		"}).Select(this2 => new {\n" +
		"    id = Id(this2.this1.this0.doc),\n" +
		"    date = this2.this1.this0.item.date,\n" +
		"    latitude = this2.this1.lat,\n" +
		"    longitude = this2.lng,\n" +
		"    coordinates = this.CreateSpatialField(((double ? ) this2.this1.lat), ((double ? ) this2.lng))\n" +
		"})"
	res.Store("id", ravendb.FieldStorageYes)
	res.Store("date", ravendb.FieldStorageYes)

	res.Store("latitude", ravendb.FieldStorageYes)
	res.Store("longitude", ravendb.FieldStorageYes)
	return res
}

func spatialWeirdSpatialResults(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		myDocument := &MyDocument{}
		myDocument.ID = "First"

		myDocumentItem := &MyDocumentItem{}
		myDocumentItem.Date = time.Now()
		myDocumentItem.Latitude = 10.0
		myDocumentItem.Longitude = 10.0

		myDocument.Items = []*MyDocumentItem{myDocumentItem}

		err = session.Store(myDocument)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	index := NewMyIndex()
	err = index.Execute(store, nil, "")
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		var statsRef *ravendb.QueryStatistics

		q := session.Advanced().QueryIndex(index.IndexName)
		q = q.WaitForNonStaleResults(0)
		q = q.WithinRadiusOf("coordinates", 0, 12.3456789, 12.3456789)
		q = q.Statistics(&statsRef)
		projType := reflect.TypeOf(&MyDocument{})
		q = q.SelectFields(projType, "id", "latitude", "longitude")
		q = q.Take(50)

		var result []*MyDocument
		err = q.GetResults(&result)
		assert.NoError(t, err)

		assert.Equal(t, statsRef.TotalResults, 0)

		assert.Equal(t, len(result), 0)

		session.Close()
	}
}

func spatialMatchSpatialResults(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		myDocument := &MyDocument{}
		myDocument.ID = "First"

		myDocumentItem := &MyDocumentItem{}
		myDocumentItem.Date = time.Now()
		myDocumentItem.Latitude = 10.0
		myDocumentItem.Longitude = 10.0

		myDocument.Items = []*MyDocumentItem{myDocumentItem}

		err = session.Store(myDocument)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	index := NewMyIndex()
	err = index.Execute(store, nil, "")
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		var statsRef *ravendb.QueryStatistics

		q := session.Advanced().QueryIndex(index.IndexName)
		q = q.WaitForNonStaleResults(0)
		q = q.WithinRadiusOf("coordinates", 0, 10, 10)
		q = q.Statistics(&statsRef)
		projType := reflect.TypeOf(&MyDocument{})
		q = q.SelectFields(projType, "id", "latitude", "longitude")
		q = q.Take(50)

		var result []*MyDocument
		err = q.GetResults(&result)
		assert.NoError(t, err)

		assert.Equal(t, statsRef.TotalResults, 1)

		assert.Equal(t, len(result), 1)

		session.Close()
	}
}

func TestSpatial(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests

	spatialWeirdSpatialResults(t, driver)
	spatialMatchSpatialResults(t, driver)
}
