package tests

import (
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func ravendb_9676_canOrderByDistanceOnDynamicSpatialField(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		item := &Item{
			Name:      "Item1",
			Latitude:  10,
			Longitude: 10,
		}

		err = session.Store(item)
		assert.NoError(t, err)

		item1 := &Item{
			Name:      "Item2",
			Latitude:  11,
			Longitude: 11,
		}

		err = session.Store(item1)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q := session.Query(ravendb.GetTypeOf(&Item{}))
		q = q.WaitForNonStaleResults(0)
		f := ravendb.NewPointField("latitude", "longitude")
		fn := func(f *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			return f.WithinRadius(1000, 10, 10)
		}

		q = q.Spatial2(f, fn)
		q2 := q.OrderByDistance(ravendb.NewPointField("latitude", "longitude"), 10, 10)
		items, err := q2.ToList()
		assert.NoError(t, err)

		assert.Equal(t, len(items), 2)
		session.Close()

		item := items[0].(*Item)
		assert.Equal(t, item.Name, "Item1")

		item = items[1].(*Item)
		assert.Equal(t, item.Name, "Item2")

		q = session.Query(ravendb.GetTypeOf(&Item{}))
		q = q.WaitForNonStaleResults(0)
		f = ravendb.NewPointField("latitude", "longitude")
		q = q.Spatial2(f, fn)
		q2 = q.OrderByDistanceDescending(ravendb.NewPointField("latitude", "longitude"), 10, 10)
		items, err = q2.ToList()

		assert.NoError(t, err)

		assert.Equal(t, len(items), 2)
		session.Close()

		item = items[0].(*Item)
		assert.Equal(t, item.Name, "Item2")

		item = items[1].(*Item)
		assert.Equal(t, item.Name, "Item1")
	}
}

type Item struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func TestRavenDB9676(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches the order of Java tests
	ravendb_9676_canOrderByDistanceOnDynamicSpatialField(t)
}
