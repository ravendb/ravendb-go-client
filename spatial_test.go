package ravendb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func NewMyIndex() *AbstractIndexCreationTask {
	res := NewAbstractIndexCreationTask("MyIndex")
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
	res.store("id", FieldStorage_YES)
	res.store("date", FieldStorage_YES)

	res.store("latitude", FieldStorage_YES)
	res.store("longitude", FieldStorage_YES)
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
	err = index.execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		var statsRef *QueryStatistics

		q := session.Advanced().DocumentQueryInIndex(getTypeOf(&MyDocument{}), index)
		q = q.waitForNonStaleResults(0)
		q = q.withinRadiusOf("coordinates", 0, 12.3456789, 12.3456789)
		q = q.statistics(&statsRef)
		q = q.selectFields(getTypeOf(&MyProjection{}), "id", "latitude", "longitude")
		q = q.take(50)

		result, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, statsRef.getTotalResults(), 0)

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
	err = index.execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		var statsRef *QueryStatistics

		q := session.Advanced().DocumentQueryInIndex(getTypeOf(&MyDocument{}), index)
		q = q.waitForNonStaleResults(0)
		q = q.withinRadiusOf("coordinates", 0, 10, 10)
		q = q.statistics(&statsRef)
		q = q.selectFields(getTypeOf(&MyProjection{}), "id", "latitude", "longitude")
		q = q.take(50)

		result, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, statsRef.getTotalResults(), 1)

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
	spatial_matchSpatialResults(t)
}
