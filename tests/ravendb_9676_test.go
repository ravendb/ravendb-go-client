package tests

import (
	"reflect"
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func ravendb_9676_canOrderByDistanceOnDynamicSpatialField(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
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

		var items []*Item
		q := session.QueryOld(reflect.TypeOf(&Item{}))
		q = q.WaitForNonStaleResults(0)
		f := ravendb.NewPointField("latitude", "longitude")
		fn := func(f *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			return f.WithinRadius(1000, 10, 10)
		}

		q = q.Spatial2(f, fn)
		q2 := q.OrderByDistance(ravendb.NewPointField("latitude", "longitude"), 10, 10)
		err = q2.ToList(&items)
		assert.NoError(t, err)

		assert.Equal(t, len(items), 2)
		session.Close()

		item := items[0]
		assert.Equal(t, item.Name, "Item1")

		item = items[1]
		assert.Equal(t, item.Name, "Item2")

		items = nil
		q = session.QueryOld(reflect.TypeOf(&Item{}))
		q = q.WaitForNonStaleResults(0)
		f = ravendb.NewPointField("latitude", "longitude")
		q = q.Spatial2(f, fn)
		q2 = q.OrderByDistanceDescending(ravendb.NewPointField("latitude", "longitude"), 10, 10)
		err = q2.ToList(&items)

		assert.NoError(t, err)

		assert.Equal(t, len(items), 2)
		session.Close()

		item = items[0]
		assert.Equal(t, item.Name, "Item2")

		item = items[1]
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

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches the order of Java tests
	ravendb_9676_canOrderByDistanceOnDynamicSpatialField(t, driver)
}
