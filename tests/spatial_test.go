package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

type MyDocumentItem struct {
	Date      time.Time `json:"date"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
}

func (i *MyDocumentItem) getDate() time.Time {
	return i.Date
}

func (i *MyDocumentItem) setDate(date time.Time) {
	i.Date = date
}

func (i *MyDocumentItem) getLatitude() float64 {
	return i.Latitude
}

func (i *MyDocumentItem) setLatitude(latitude float64) {
	i.Latitude = latitude
}

func (i *MyDocumentItem) getLongitude() float64 {
	return i.Longitude
}

func (i *MyDocumentItem) setLongitude(longitude float64) {
	i.Longitude = longitude
}

type MyDocument struct {
	ID    string
	Items []*MyDocumentItem `json:"items"`
}

func (d *MyDocument) getId() string {
	return d.ID
}

func (d *MyDocument) setId(id string) {
	d.ID = id
}

func (d *MyDocument) getItems() []*MyDocumentItem {
	return d.Items
}

func (d *MyDocument) setItems(items []*MyDocumentItem) {
	d.Items = items
}

type MyProjection struct {
	ID        string
	Date      time.Time `json:"date"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
}

func (p *MyProjection) getId() string {
	return p.ID
}

func (p *MyProjection) setId(id string) {
	p.ID = id
}

func (p *MyProjection) getDate() time.Time {
	return p.Date
}

func (p *MyProjection) setDate(date time.Time) {
	p.Date = date
}

func (p *MyProjection) getLatitude() float64 {
	return p.Latitude
}

func (p *MyProjection) setLatitude(latitude float64) {
	p.Latitude = latitude
}

func (p *MyProjection) getLongitude() float64 {
	return p.Longitude
}

func (p *MyProjection) setLongitude(longitude float64) {
	p.Longitude = longitude
}

func NewMyIndex() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("MyIndex")
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
	res.Store("id", ravendb.FieldStorage_YES)
	res.Store("date", ravendb.FieldStorage_YES)

	res.Store("latitude", ravendb.FieldStorage_YES)
	res.Store("longitude", ravendb.FieldStorage_YES)
	return res
}

func spatial_weirdSpatialResults(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		myDocument := &MyDocument{}
		myDocument.setId("First")

		myDocumentItem := &MyDocumentItem{}
		myDocumentItem.setDate(time.Now())
		myDocumentItem.setLatitude(10.0)
		myDocumentItem.setLongitude(10.0)

		myDocument.setItems([]*MyDocumentItem{myDocumentItem})

		err = session.Store(myDocument)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	index := NewMyIndex()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		var statsRef *ravendb.QueryStatistics

		q := session.Advanced().DocumentQueryInIndexOld(ravendb.GetTypeOf(&MyDocument{}), index)
		q = q.WaitForNonStaleResults(0)
		q = q.WithinRadiusOf("coordinates", 0, 12.3456789, 12.3456789)
		q = q.Statistics(&statsRef)
		q = q.SelectFields(ravendb.GetTypeOf(&MyProjection{}), "id", "latitude", "longitude")
		q = q.Take(50)

		result, err := q.ToListOld()
		assert.NoError(t, err)

		assert.Equal(t, statsRef.GetTotalResults(), 0)

		assert.Equal(t, len(result), 0)

		session.Close()
	}
}

func spatial_matchSpatialResults(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		myDocument := &MyDocument{}
		myDocument.setId("First")

		myDocumentItem := &MyDocumentItem{}
		myDocumentItem.setDate(time.Now())
		myDocumentItem.setLatitude(10.0)
		myDocumentItem.setLongitude(10.0)

		myDocument.setItems([]*MyDocumentItem{myDocumentItem})

		err = session.Store(myDocument)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	index := NewMyIndex()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		var statsRef *ravendb.QueryStatistics

		q := session.Advanced().DocumentQueryInIndexOld(ravendb.GetTypeOf(&MyDocument{}), index)
		q = q.WaitForNonStaleResults(0)
		q = q.WithinRadiusOf("coordinates", 0, 10, 10)
		q = q.Statistics(&statsRef)
		q = q.SelectFields(ravendb.GetTypeOf(&MyProjection{}), "id", "latitude", "longitude")
		q = q.Take(50)

		result, err := q.ToListOld()
		assert.NoError(t, err)

		assert.Equal(t, statsRef.GetTotalResults(), 1)

		assert.Equal(t, len(result), 1)

		session.Close()
	}
}

func TestSpatial(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	spatial_weirdSpatialResults(t)
	if ravendb.EnableFlakyTests {
		// is flaky on CI e.g. https://travis-ci.org/kjk/ravendb-go-client/builds/416175659?utm_source=email&utm_medium=notification
		// works on my mak
		spatial_matchSpatialResults(t)
	}
}
