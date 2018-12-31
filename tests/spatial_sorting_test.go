package tests

import (
	"fmt"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

const (
	FilteredLat    = float64(44.419575)
	FilteredLng    = float64(34.042618)
	SortedLat      = float64(44.417398)
	SortedLng      = float64(34.042575)
	FilteredRadius = float64(100)
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

func spatialSortingCreateData(t *testing.T, driver *RavenTestDriver, store *ravendb.IDocumentStore) {
	var err error
	indexDefinition := ravendb.NewIndexDefinition()
	indexDefinition.Name = "eventsByLatLng"
	indexDefinition.Maps = []string{"from e in docs.Shops select new { e.venue, coordinates = CreateSpatialField(e.latitude, e.longitude) }"}

	fields := make(map[string]*ravendb.IndexFieldOptions)
	options := ravendb.NewIndexFieldOptions()
	options.Indexing = ravendb.FieldIndexingExact
	fields["tag"] = options
	indexDefinition.Fields = fields

	op := ravendb.NewPutIndexesOperation(indexDefinition)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)

	indexDefinition2 := ravendb.NewIndexDefinition()
	indexDefinition2.Name = "eventsByLatLngWSpecialField"
	indexDefinition2.Maps = []string{"from e in docs.Shops select new { e.venue, mySpacialField = CreateSpatialField(e.latitude, e.longitude) }"}

	indexFieldOptions := ravendb.NewIndexFieldOptions()
	indexFieldOptions.Indexing = ravendb.FieldIndexingExact
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
	ok := stringArrayContainsExactly(resultIDs, expectedOrder)
	assert.True(t, ok)
}

func spatialSortingCanFilterByLocationAndSortByDistanceFromDifferentPointWDocQuery(t *testing.T,
	driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	spatialSortingCreateData(t, driver, store)

	{
		session := openSessionMust(t, store)

		var shops []*Shop
		queryIndex := &ravendb.Query{
			IndexName: "eventsByLatLng",
		}
		q := session.QueryWithQuery(queryIndex)
		fn := func(f *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			res := f.Within(getQueryShapeFromLatLon(FilteredLat, FilteredLng, FilteredRadius))
			return res
		}

		q = q.Spatial3("coordinates", fn)
		q = q.OrderByDistanceLatLong("coordinates", SortedLat, SortedLng)
		err = q.GetResults(&shops)
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

func spatialSortingCanSortByDistanceWOFilteringWDocQuery(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	spatialSortingCreateData(t, driver, store)

	{
		session := openSessionMust(t, store)

		var shops []*Shop
		queryIndex := &ravendb.Query{
			IndexName: "eventsByLatLng",
		}
		q := session.QueryWithQuery(queryIndex)
		q = q.OrderByDistanceLatLong("coordinates", SortedLat, SortedLng)

		err := q.GetResults(&shops)
		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(sortedExpectedOrder))

		ids := getShopIDs(shops)
		assertResultsOrder(t, ids, sortedExpectedOrder)

		session.Close()
	}
}

func spatialSortingCanSortByDistanceWOFilteringWDocQueryBySpecifiedField(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	spatialSortingCreateData(t, driver, store)

	{
		session := openSessionMust(t, store)

		var shops []*Shop
		queryIndex := &ravendb.Query{
			IndexName: "eventsByLatLngWSpecialField",
		}
		q := session.QueryWithQuery(queryIndex)
		q = q.OrderByDistanceLatLong("mySpacialField", SortedLat, SortedLng)
		err := q.GetResults(&shops)
		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(sortedExpectedOrder))

		ids := getShopIDs(shops)
		assertResultsOrder(t, ids, sortedExpectedOrder)

		session.Close()
	}
}

func spatialSortingCanSortByDistanceWOFiltering(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	spatialSortingCreateData(t, driver, store)

	{
		session := openSessionMust(t, store)

		var shops []*Shop
		queryIndex := &ravendb.Query{
			IndexName: "eventsByLatLng",
		}
		q := session.QueryWithQuery(queryIndex)
		q = q.OrderByDistanceLatLong("coordinates", FilteredLat, FilteredLng)
		err := q.GetResults(&shops)

		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(filteredExpectedOrder))

		ids := getShopIDs(shops)
		assertResultsOrder(t, ids, filteredExpectedOrder)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var shops []*Shop
		queryIndex := &ravendb.Query{
			IndexName: "eventsByLatLng",
		}
		q := session.QueryWithQuery(queryIndex)
		q = q.OrderByDistanceDescendingLatLong("coordinates", FilteredLat, FilteredLng)
		err := q.GetResults(&shops)

		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(filteredExpectedOrder))

		ids := getShopIDs(shops)
		stringArrayReverse(ids)
		assertResultsOrder(t, ids, filteredExpectedOrder)

		session.Close()
	}
}

func spatialSortingCanSortByDistanceWOFilteringBySpecifiedField(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	spatialSortingCreateData(t, driver, store)

	{
		session := openSessionMust(t, store)

		var shops []*Shop
		queryIndex := &ravendb.Query{
			IndexName: "eventsByLatLngWSpecialField",
		}
		q := session.QueryWithQuery(queryIndex)
		q = q.OrderByDistanceLatLong("mySpacialField", FilteredLat, FilteredLng)
		err := q.GetResults(&shops)

		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(filteredExpectedOrder))

		ids := getShopIDs(shops)
		assertResultsOrder(t, ids, filteredExpectedOrder)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var shops []*Shop
		queryIndex := &ravendb.Query{
			IndexName: "eventsByLatLngWSpecialField",
		}
		q := session.QueryWithQuery(queryIndex)
		q = q.OrderByDistanceDescendingLatLong("mySpacialField", FilteredLat, FilteredLng)
		err := q.GetResults(&shops)

		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(filteredExpectedOrder))

		ids := getShopIDs(shops)
		stringArrayReverse(ids)
		assertResultsOrder(t, ids, filteredExpectedOrder)

		session.Close()
	}
}

func getQueryShapeFromLatLon(lat float64, lng float64, radius float64) string {
	return fmt.Sprintf("Circle(%f %f d=%f)", lng, lat, radius)
}

func TestSpatialSorting(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	spatialSortingCanSortByDistanceWOFilteringBySpecifiedField(t, driver)
	spatialSortingCanFilterByLocationAndSortByDistanceFromDifferentPointWDocQuery(t, driver)
	spatialSortingCanSortByDistanceWOFiltering(t, driver)
	spatialSortingCanSortByDistanceWOFilteringWDocQuery(t, driver)
	spatialSortingCanSortByDistanceWOFilteringWDocQueryBySpecifiedField(t, driver)
}
