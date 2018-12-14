package tests

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

const (
	FILTERED_LAT    = float64(44.419575)
	FILTERED_LNG    = float64(34.042618)
	SORTED_LAT      = float64(44.417398)
	SORTED_LNG      = float64(34.042575)
	FILTERED_RADIUS = float64(100)
)

type Shop struct {
	ID        string
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func NewShop(latitude float64, longitude float64) *Shop {
	return &Shop{
		Latitude:  latitude,
		Longitude: longitude,
	}
}

var (
	shops = []*Shop{
		NewShop(44.420678, 34.042490),
		NewShop(44.419712, 34.042232),
		NewShop(44.418686, 34.043219),
	}
	//shop/1:0.36KM, shop/2:0.26KM, shop/3 0.15KM from (34.042575,  44.417398)
	sortedExpectedOrder = []string{"shops/3-A", "shops/2-A", "shops/1-A"}

	//shop/1:0.12KM, shop/2:0.03KM, shop/3 0.11KM from (34.042618,  44.419575)
	filteredExpectedOrder = []string{"shops/2-A", "shops/3-A", "shops/1-A"}
)

func spatialSorting_createData(t *testing.T, driver *RavenTestDriver, store *ravendb.IDocumentStore) {
	var err error
	indexDefinition := ravendb.NewIndexDefinition()
	indexDefinition.Name = "eventsByLatLng"
	indexDefinition.Maps = []string{"from e in docs.Shops select new { e.venue, coordinates = CreateSpatialField(e.latitude, e.longitude) }"}

	fields := make(map[string]*ravendb.IndexFieldOptions)
	options := ravendb.NewIndexFieldOptions()
	options.Indexing = ravendb.FieldIndexing_EXACT
	fields["tag"] = options
	indexDefinition.Fields = fields

	op := ravendb.NewPutIndexesOperation(indexDefinition)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)

	indexDefinition2 := ravendb.NewIndexDefinition()
	indexDefinition2.Name = "eventsByLatLngWSpecialField"
	indexDefinition2.Maps = []string{"from e in docs.Shops select new { e.venue, mySpacialField = CreateSpatialField(e.latitude, e.longitude) }"}

	indexFieldOptions := ravendb.NewIndexFieldOptions()
	indexFieldOptions.Indexing = ravendb.FieldIndexing_EXACT
	fields = map[string]*ravendb.IndexFieldOptions{
		"tag": indexFieldOptions,
	}
	indexDefinition2.Fields = fields

	op = ravendb.NewPutIndexesOperation(indexDefinition2)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		for _, shop := range shops {
			err = session.Store(shop)
			assert.NoError(t, err)
		}

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)
}

func assertResultsOrder(t *testing.T, resultIDs []string, expectedOrder []string) {
	ok := ravendb.StringArrayContainsExactly(resultIDs, expectedOrder)
	assert.True(t, ok)
}

func spatialSorting_canFilterByLocationAndSortByDistanceFromDifferentPointWDocQuery(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	spatialSorting_createData(t, driver, store)

	{
		session := openSessionMust(t, store)

		var shops []*Shop
		q := session.QueryWithQueryOld(reflect.TypeOf(&Shop{}), ravendb.Query_index("eventsByLatLng"))
		fn := func(f *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			res := f.Within(getQueryShapeFromLatLon(FILTERED_LAT, FILTERED_LNG, FILTERED_RADIUS))
			return res
		}

		q = q.Spatial3("coordinates", fn)
		q = q.OrderByDistanceLatLong("coordinates", SORTED_LAT, SORTED_LNG)
		err = q.ToList(&shops)
		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(sortedExpectedOrder))

		ids := getShopIDs(shops)
		assertResultsOrder(t, ids, sortedExpectedOrder)

		session.Close()
	}
}

func getShopIDs(shops []*Shop) []string {
	var res []string
	for _, shop := range shops {
		id := shop.ID
		res = append(res, id)
	}
	return res
}

func spatialSorting_canSortByDistanceWOFilteringWDocQuery(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	spatialSorting_createData(t, driver, store)

	{
		session := openSessionMust(t, store)

		var shops []*Shop
		q := session.QueryWithQueryOld(reflect.TypeOf(&Shop{}), ravendb.Query_index("eventsByLatLng"))
		q = q.OrderByDistanceLatLong("coordinates", SORTED_LAT, SORTED_LNG)

		err := q.ToList(&shops)
		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(sortedExpectedOrder))

		ids := getShopIDs(shops)
		assertResultsOrder(t, ids, sortedExpectedOrder)

		session.Close()
	}
}

func spatialSorting_canSortByDistanceWOFilteringWDocQueryBySpecifiedField(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	spatialSorting_createData(t, driver, store)

	{
		session := openSessionMust(t, store)

		var shops []*Shop
		q := session.QueryWithQueryOld(reflect.TypeOf(&Shop{}), ravendb.Query_index("eventsByLatLngWSpecialField"))
		q = q.OrderByDistanceLatLong("mySpacialField", SORTED_LAT, SORTED_LNG)
		err := q.ToList(&shops)
		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(sortedExpectedOrder))

		ids := getShopIDs(shops)
		assertResultsOrder(t, ids, sortedExpectedOrder)

		session.Close()
	}
}

func spatialSorting_canSortByDistanceWOFiltering(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	spatialSorting_createData(t, driver, store)

	{
		session := openSessionMust(t, store)

		var shops []*Shop
		q := session.QueryWithQueryOld(reflect.TypeOf(&Shop{}), ravendb.Query_index("eventsByLatLng"))
		q = q.OrderByDistanceLatLong("coordinates", FILTERED_LAT, FILTERED_LNG)
		err := q.ToList(&shops)

		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(filteredExpectedOrder))

		ids := getShopIDs(shops)
		assertResultsOrder(t, ids, filteredExpectedOrder)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var shops []*Shop
		q := session.QueryWithQueryOld(reflect.TypeOf(&Shop{}), ravendb.Query_index("eventsByLatLng"))
		q = q.OrderByDistanceDescendingLatLong("coordinates", FILTERED_LAT, FILTERED_LNG)
		err := q.ToList(&shops)

		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(filteredExpectedOrder))

		ids := getShopIDs(shops)
		ravendb.StringArrayReverse(ids)
		assertResultsOrder(t, ids, filteredExpectedOrder)

		session.Close()
	}
}

func spatialSorting_canSortByDistanceWOFilteringBySpecifiedField(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	spatialSorting_createData(t, driver, store)

	{
		session := openSessionMust(t, store)

		var shops []*Shop
		q := session.QueryWithQueryOld(reflect.TypeOf(&Shop{}), ravendb.Query_index("eventsByLatLngWSpecialField"))
		q = q.OrderByDistanceLatLong("mySpacialField", FILTERED_LAT, FILTERED_LNG)
		err := q.ToList(&shops)

		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(filteredExpectedOrder))

		ids := getShopIDs(shops)
		assertResultsOrder(t, ids, filteredExpectedOrder)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var shops []*Shop
		q := session.QueryWithQueryOld(reflect.TypeOf(&Shop{}), ravendb.Query_index("eventsByLatLngWSpecialField"))
		q = q.OrderByDistanceDescendingLatLong("mySpacialField", FILTERED_LAT, FILTERED_LNG)
		err := q.ToList(&shops)

		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(filteredExpectedOrder))

		ids := getShopIDs(shops)
		ravendb.StringArrayReverse(ids)
		assertResultsOrder(t, ids, filteredExpectedOrder)

		session.Close()
	}
}

func getQueryShapeFromLatLon(lat float64, lng float64, radius float64) string {
	return fmt.Sprintf("Circle(%f %f d=%f)", lng, lat, radius)
}

func TestSpatialSorting(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	spatialSorting_canSortByDistanceWOFilteringBySpecifiedField(t, driver)
	spatialSorting_canFilterByLocationAndSortByDistanceFromDifferentPointWDocQuery(t, driver)
	spatialSorting_canSortByDistanceWOFiltering(t, driver)
	spatialSorting_canSortByDistanceWOFilteringWDocQuery(t, driver)
	spatialSorting_canSortByDistanceWOFilteringWDocQueryBySpecifiedField(t, driver)
}
