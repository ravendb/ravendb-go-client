package ravendb

import (
	"fmt"
	"testing"

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

func (s *Shop) getId() string {
	return s.ID
}

func (s *Shop) setId(id string) {
	s.ID = id
}

func (s *Shop) getLatitude() float64 {
	return s.Latitude
}

func (s *Shop) setLatitude(latitude float64) {
	s.Latitude = latitude
}

func (s *Shop) getLongitude() float64 {
	return s.Longitude
}

func (s *Shop) setLongitude(longitude float64) {
	s.Longitude = longitude
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

func spatialSorting_createData(t *testing.T, store *IDocumentStore) {
	var err error
	indexDefinition := NewIndexDefinition()
	indexDefinition.setName("eventsByLatLng")
	maps := NewStringSetFromStrings("from e in docs.Shops select new { e.venue, coordinates = CreateSpatialField(e.latitude, e.longitude) }")
	indexDefinition.setMaps(maps)

	fields := make(map[string]*IndexFieldOptions)
	options := NewIndexFieldOptions()
	options.setIndexing(FieldIndexing_EXACT)
	fields["tag"] = options
	indexDefinition.setFields(fields)

	op := NewPutIndexesOperation(indexDefinition)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)

	indexDefinition2 := NewIndexDefinition()
	indexDefinition2.setName("eventsByLatLngWSpecialField")
	maps = NewStringSetFromStrings("from e in docs.Shops select new { e.venue, mySpacialField = CreateSpatialField(e.latitude, e.longitude) }")
	indexDefinition2.setMaps(maps)

	indexFieldOptions := NewIndexFieldOptions()
	indexFieldOptions.setIndexing(FieldIndexing_EXACT)
	fields = map[string]*IndexFieldOptions{
		"tag": indexFieldOptions,
	}
	indexDefinition2.setFields(fields)

	op = NewPutIndexesOperation(indexDefinition2)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		for _, shop := range shops {
			err = session.Store(shop)
			assert.NoError(t, err)
		}

		err = session.SaveChanges()

		session.Close()
	}

	err = gRavenTestDriver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)
}

func assertResultsOrder(t *testing.T, resultIDs []string, expectedOrder []string) {
	ok := stringArrayContainsExactly(resultIDs, expectedOrder)
	assert.True(t, ok)
}

func spatialSorting_canFilterByLocationAndSortByDistanceFromDifferentPointWDocQuery(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	spatialSorting_createData(t, store)

	{
		session := openSessionMust(t, store)

		q := session.QueryWithQuery(GetTypeOf(&Shop{}), Query_index("eventsByLatLng"))
		fn := func(f *SpatialCriteriaFactory) SpatialCriteria {
			res := f.within(getQueryShapeFromLatLon(FILTERED_LAT, FILTERED_LNG, FILTERED_RADIUS))
			return res
		}

		q = q.Spatial3("coordinates", fn)
		q = q.OrderByDistanceLatLong("coordinates", SORTED_LAT, SORTED_LNG)
		shops, err := q.ToList()
		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(sortedExpectedOrder))

		ids := getShopIDs(shops)
		assertResultsOrder(t, ids, sortedExpectedOrder)

		session.Close()
	}
}

func getShopIDs(shops []interface{}) []string {
	var res []string
	for _, el := range shops {
		shop := el.(*Shop)
		id := shop.getId()
		res = append(res, id)
	}
	return res
}

func spatialSorting_canSortByDistanceWOFilteringWDocQuery(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	spatialSorting_createData(t, store)

	{
		session := openSessionMust(t, store)

		q := session.QueryWithQuery(GetTypeOf(&Shop{}), Query_index("eventsByLatLng"))
		q = q.OrderByDistanceLatLong("coordinates", SORTED_LAT, SORTED_LNG)

		shops, err := q.ToList()
		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(sortedExpectedOrder))

		ids := getShopIDs(shops)
		assertResultsOrder(t, ids, sortedExpectedOrder)

		session.Close()
	}
}

func spatialSorting_canSortByDistanceWOFilteringWDocQueryBySpecifiedField(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	spatialSorting_createData(t, store)

	{
		session := openSessionMust(t, store)

		q := session.QueryWithQuery(GetTypeOf(&Shop{}), Query_index("eventsByLatLngWSpecialField"))
		q = q.OrderByDistanceLatLong("mySpacialField", SORTED_LAT, SORTED_LNG)
		shops, err := q.ToList()
		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(sortedExpectedOrder))

		ids := getShopIDs(shops)
		assertResultsOrder(t, ids, sortedExpectedOrder)

		session.Close()
	}
}

func spatialSorting_canSortByDistanceWOFiltering(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	spatialSorting_createData(t, store)

	{
		session := openSessionMust(t, store)
		q := session.QueryWithQuery(GetTypeOf(&Shop{}), Query_index("eventsByLatLng"))
		q = q.OrderByDistanceLatLong("coordinates", FILTERED_LAT, FILTERED_LNG)
		shops, err := q.ToList()

		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(filteredExpectedOrder))

		ids := getShopIDs(shops)
		assertResultsOrder(t, ids, filteredExpectedOrder)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q := session.QueryWithQuery(GetTypeOf(&Shop{}), Query_index("eventsByLatLng"))
		q = q.orderByDistanceDescendingLatLong("coordinates", FILTERED_LAT, FILTERED_LNG)
		shops, err := q.ToList()

		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(filteredExpectedOrder))

		ids := getShopIDs(shops)
		stringArrayReverse(ids)
		assertResultsOrder(t, ids, filteredExpectedOrder)

		session.Close()
	}
}

func spatialSorting_canSortByDistanceWOFilteringBySpecifiedField(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	spatialSorting_createData(t, store)

	{
		session := openSessionMust(t, store)

		q := session.QueryWithQuery(GetTypeOf(&Shop{}), Query_index("eventsByLatLngWSpecialField"))
		q = q.OrderByDistanceLatLong("mySpacialField", FILTERED_LAT, FILTERED_LNG)
		shops, err := q.ToList()

		assert.NoError(t, err)
		assert.Equal(t, len(shops), len(filteredExpectedOrder))

		ids := getShopIDs(shops)
		assertResultsOrder(t, ids, filteredExpectedOrder)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q := session.QueryWithQuery(GetTypeOf(&Shop{}), Query_index("eventsByLatLngWSpecialField"))
		q = q.orderByDistanceDescendingLatLong("mySpacialField", FILTERED_LAT, FILTERED_LNG)
		shops, err := q.ToList()

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
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	spatialSorting_canSortByDistanceWOFilteringBySpecifiedField(t)
	spatialSorting_canFilterByLocationAndSortByDistanceFromDifferentPointWDocQuery(t)
	spatialSorting_canSortByDistanceWOFiltering(t)
	spatialSorting_canSortByDistanceWOFilteringWDocQuery(t)
	spatialSorting_canSortByDistanceWOFilteringWDocQueryBySpecifiedField(t)
}
