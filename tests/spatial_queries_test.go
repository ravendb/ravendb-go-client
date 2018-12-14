package tests

import (
	"reflect"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func NewSpatialQueriesInMemoryTestIdx() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("SpatialQueriesInMemoryTestIdx")
	res.Map = "docs.Listings.Select(listingItem => new {\n" +
		"    classCodes = listingItem.classCodes,\n" +
		"    latitude = listingItem.latitude,\n" +
		"    longitude = listingItem.longitude,\n" +
		"    coordinates = this.CreateSpatialField(((double ? )((double)(listingItem.latitude))), ((double ? )((double)(listingItem.longitude))))\n" +
		"})"
	return res
}

func spatialQueries_canRunSpatialQueriesInMemory(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	err = NewSpatialQueriesInMemoryTestIdx().Execute(store)
	assert.NoError(t, err)
}

type Listing struct {
	ClassCodes string `json:"classCodes"`
	Latitude   int64  `json:"latitude"`
	Longitude  int64  `json:"longitude"`
}

func (l *Listing) getClassCodes() string {
	return l.ClassCodes
}

func (l *Listing) setClassCodes(classCodes string) {
	l.ClassCodes = classCodes
}

func (l *Listing) getLatitude() int64 {
	return l.Latitude
}

func (l *Listing) setLatitude(latitude int64) {
	l.Latitude = latitude
}

func (l *Listing) getLongitude() int64 {
	return l.Longitude
}

func (l *Listing) setLongitude(longitude int64) {
	l.Longitude = longitude
}

func spatialQueries_canSuccessfullyDoSpatialQueryOfNearbyLocations(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	areaOneDocOne := NewDummyGeoDoc(55.6880508001, 13.5717346673)
	areaOneDocTwo := NewDummyGeoDoc(55.6821978456, 13.6076183965)
	areaOneDocThree := NewDummyGeoDoc(55.673251569, 13.5946697607)

	// This item is 12 miles (approx 19 km) from the closest in areaOne
	closeButOutsideAreaOne := NewDummyGeoDoc(55.8634157297, 13.5497731987)

	// This item is about 3900 miles from areaOne
	newYork := NewDummyGeoDoc(40.7137578228, -74.0126901936)

	{
		session := openSessionMust(t, store)

		err = session.Store(areaOneDocOne)
		assert.NoError(t, err)
		err = session.Store(areaOneDocTwo)
		assert.NoError(t, err)
		err = session.Store(areaOneDocThree)
		assert.NoError(t, err)
		err = session.Store(closeButOutsideAreaOne)
		assert.NoError(t, err)
		err = session.Store(newYork)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		indexDefinition := ravendb.NewIndexDefinition()
		indexDefinition.Name = "FindByLatLng"
		indexDefinition.Maps = []string{"from doc in docs select new { coordinates = CreateSpatialField(doc.latitude, doc.longitude) }"}

		op := ravendb.NewPutIndexesOperation(indexDefinition)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)

		// Wait until the index is built
		var notUsed []*DummyGeoDoc
		q := session.QueryWithQueryOld(reflect.TypeOf(&DummyGeoDoc{}), ravendb.Query_index("FindByLatLng"))
		q = q.WaitForNonStaleResults(0)
		err = q.ToList(&notUsed)
		assert.NoError(t, err)

		lat := float64(55.6836422426)
		lng := float64(13.5871808352) // in the middle of AreaOne
		radius := float64(5.0)

		var nearbyDocs []*DummyGeoDoc
		q = session.QueryWithQueryOld(reflect.TypeOf(&DummyGeoDoc{}), ravendb.Query_index("FindByLatLng"))
		q = q.WithinRadiusOf("coordinates", radius, lat, lng)
		q = q.WaitForNonStaleResults(0)
		err = q.ToList(&nearbyDocs)
		assert.NoError(t, err)

		assert.Equal(t, len(nearbyDocs), 3)

		session.Close()
	}
}

func spatialQueries_canSuccessfullyQueryByMiles(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	myHouse := NewDummyGeoDoc(44.757767, -93.355322)

	// The gym is about 7.32 miles (11.79 kilometers) from my house.
	gym := NewDummyGeoDoc(44.682861, -93.25)
	{
		session := openSessionMust(t, store)

		err = session.Store(myHouse)
		assert.NoError(t, err)
		err = session.Store(gym)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		indexDefinition := ravendb.NewIndexDefinition()
		indexDefinition.Name = "FindByLatLng"
		indexDefinition.Maps = []string{"from doc in docs select new { coordinates = CreateSpatialField(doc.latitude, doc.longitude) }"}

		op := ravendb.NewPutIndexesOperation(indexDefinition)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)

		// Wait until the index is built
		var notUsed []*DummyGeoDoc
		q := session.QueryWithQueryOld(reflect.TypeOf(&DummyGeoDoc{}), ravendb.Query_index("FindByLatLng"))
		q = q.WaitForNonStaleResults(0)
		err = q.ToList(&notUsed)
		assert.NoError(t, err)

		radius := float64(8)

		// Find within 8 miles.
		// We should find both my house and the gym.

		var matchesWithinMiles []*DummyGeoDoc
		q = session.QueryWithQueryOld(reflect.TypeOf(&DummyGeoDoc{}), ravendb.Query_index("FindByLatLng"))
		q = q.WithinRadiusOfWithUnits("coordinates", radius, myHouse.Latitude, myHouse.Longitude, ravendb.SpatialUnits_MILES)
		q = q.WaitForNonStaleResults(0)
		err = q.ToList(&matchesWithinMiles)
		assert.NoError(t, err)
		assert.Equal(t, len(matchesWithinMiles), 2)

		// Find within 8 kilometers.
		// We should find only my house, since the gym is ~11 kilometers out.

		var matchesWithinKilometers []*DummyGeoDoc
		q = session.QueryWithQueryOld(reflect.TypeOf(&DummyGeoDoc{}), ravendb.Query_index("FindByLatLng"))
		q = q.WithinRadiusOfWithUnits("coordinates", radius, myHouse.Latitude, myHouse.Longitude, ravendb.SpatialUnits_KILOMETERS)
		q = q.WaitForNonStaleResults(0)
		err = q.ToList(&matchesWithinKilometers)
		assert.NoError(t, err)
		assert.Equal(t, len(matchesWithinKilometers), 1)

		session.Close()
	}
}

type DummyGeoDoc struct {
	ID        string
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func NewDummyGeoDoc(latitude float64, longitude float64) *DummyGeoDoc {
	return &DummyGeoDoc{
		Latitude:  latitude,
		Longitude: longitude,
	}
}

func TestSpatialQueries(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	spatialQueries_canRunSpatialQueriesInMemory(t, driver)
	spatialQueries_canSuccessfullyQueryByMiles(t, driver)
	spatialQueries_canSuccessfullyDoSpatialQueryOfNearbyLocations(t, driver)
}
