package tests

import (
	"reflect"
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

// Note: renamed to Item2 to avoid conflicts
type Item2 struct {
	ID         string
	Name       string  `json:"name"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Latitude2  float64 `json:"latitude2"`
	Longitude2 float64 `json:"longitude2"`
	ShapeWkt   string  `json:"shapeWkt"`
}

func ravenDB8328_spatialOnAutoIndex(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		item := &Item2{
			Latitude:   10,
			Longitude:  20,
			Latitude2:  10,
			Longitude2: 20,
			ShapeWkt:   "POINT(20 10)",
			Name:       "Name1",
		}

		err = session.Store(item)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		clazz := reflect.TypeOf(&Item2{})
		q := session.QueryCollectionForType(clazz)
		fieldName := ravendb.NewPointField("latitude", "longitude")
		clause := func(f *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			return f.WithinRadius(10, 10, 20)
		}
		q = q.Spatial2(fieldName, clause)

		iq, err := q.GetIndexQuery()
		assert.NoError(t, err)
		assert.Equal(t, iq.GetQuery(), "from Item2s where spatial.within(spatial.point(latitude, longitude), spatial.circle($p0, $p1, $p2))")

		q = session.QueryCollectionForType(clazz)
		fieldName2 := ravendb.NewWktField("shapeWkt")
		q = q.Spatial2(fieldName2, clause)

		iq, err = q.GetIndexQuery()
		assert.NoError(t, err)
		assert.Equal(t, iq.GetQuery(), "from Item2s where spatial.within(spatial.wkt(shapeWkt), spatial.circle($p0, $p1, $p2))")

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var statsRef *ravendb.QueryStatistics

		var results []*Item2
		q := session.QueryCollectionForType(reflect.TypeOf(&Item2{}))
		q = q.Statistics(&statsRef)
		fieldName := ravendb.NewPointField("latitude", "longitude")
		clause := func(f *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			return f.WithinRadius(10, 10, 20)
		}
		q = q.Spatial2(fieldName, clause)
		err = q.GetResults(&results)
		assert.NoError(t, err)

		assert.Equal(t, len(results), 1)

		assert.Equal(t, statsRef.IndexName, "Auto/Item2s/BySpatial.point(latitude|longitude)")

		session.Close()

		statsRef = nil
		results = nil

		q = session.QueryCollectionForType(reflect.TypeOf(&Item2{}))
		q = q.Statistics(&statsRef)
		fieldName = ravendb.NewPointField("latitude2", "longitude2")
		q = q.Spatial2(fieldName, clause)
		err = q.GetResults(&results)
		assert.NoError(t, err)

		assert.Equal(t, len(results), 1)

		assert.Equal(t, statsRef.IndexName, "Auto/Item2s/BySpatial.point(latitude|longitude)AndSpatial.point(latitude2|longitude2)")

		statsRef = nil
		results = nil

		q = session.QueryCollectionForType(reflect.TypeOf(&Item2{}))
		q = q.Statistics(&statsRef)
		fieldName2 := ravendb.NewWktField("shapeWkt")
		q = q.Spatial2(fieldName2, clause)
		err = q.GetResults(&results)
		assert.NoError(t, err)

		assert.Equal(t, len(results), 1)

		assert.Equal(t, statsRef.IndexName, "Auto/Item2s/BySpatial.point(latitude|longitude)AndSpatial.point(latitude2|longitude2)AndSpatial.wkt(shapeWkt)")
	}
}

func TestRavenDB8328(t *testing.T) {
	// // t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	ravenDB8328_spatialOnAutoIndex(t, driver)
}
