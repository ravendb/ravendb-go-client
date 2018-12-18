package tests

import (
	"reflect"
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func simonBartlett_lineStringsShouldIntersect(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	index := NewGeoIndex()
	err = store.ExecuteIndex(index)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		geoDocument := &GeoDocument{
			Wkt: "LINESTRING (0 0, 1 1, 2 1)",
		}
		err = session.Store(geoDocument)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	driver.waitForIndexing(store, "", 0)

	{
		session := openSessionMust(t, store)

		// TODO: does it matter what type we send?
		q := session.QueryInIndexOld(reflect.TypeOf(&GeoDocument{}), index)
		fn := func(f *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			return f.RelatesToShape("LINESTRING (1 0, 1 1, 1 2)", ravendb.SpatialRelation_INTERSECTS)
		}
		q = q.Spatial3("WKT", fn)
		q = q.WaitForNonStaleResults(0)
		count, err := q.Count()
		assert.NoError(t, err)

		assert.Equal(t, count, 1)

		// TODO: does it matter what type we send?
		q = session.QueryInIndexOld(reflect.TypeOf(&GeoDocument{}), index)
		q = q.RelatesToShape("WKT", "LINESTRING (1 0, 1 1, 1 2)", ravendb.SpatialRelation_INTERSECTS)
		q = q.WaitForNonStaleResults(0)
		count, err = q.Count()
		assert.NoError(t, err)

		assert.Equal(t, count, 1)

		session.Close()
	}
}

func simonBartlett_circlesShouldNotIntersect(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	index := NewGeoIndex()
	err = store.ExecuteIndex(index)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		// 110km is approximately 1 degree
		geoDocument := &GeoDocument{
			Wkt: "CIRCLE(0.000000 0.000000 d=110)",
		}
		err = session.Store(geoDocument)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	driver.waitForIndexing(store, "", 0)

	{
		session := openSessionMust(t, store)

		// Should not intersect, as there is 1 Degree between the two shapes
		q := session.QueryInIndexOld(reflect.TypeOf(&GeoDocument{}), index)
		fn := func(f *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			return f.RelatesToShape("CIRCLE(0.000000 3.000000 d=110)", ravendb.SpatialRelation_INTERSECTS)
		}

		q = q.Spatial3("WKT", fn)

		q = q.WaitForNonStaleResults(0)
		count, err := q.Count()
		assert.NoError(t, err)

		assert.Equal(t, count, 0)

		q = session.QueryInIndexOld(reflect.TypeOf(&GeoDocument{}), index)
		q = q.RelatesToShape("WKT", "CIRCLE(0.000000 3.000000 d=110)", ravendb.SpatialRelation_INTERSECTS)
		q = q.WaitForNonStaleResults(0)
		count, err = q.Count()
		assert.NoError(t, err)

		assert.Equal(t, count, 0)

		session.Close()
	}
}

type GeoDocument struct {
	Wkt string `json:"WKT"`
}

func NewGeoIndex() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("GeoIndex")
	res.Map = "docs.GeoDocuments.Select(doc => new {\n" +
		"    WKT = this.CreateSpatialField(doc.WKT)\n" +
		"})"
	spatialOptions := ravendb.NewSpatialOptions()
	spatialOptions.Strategy = ravendb.SpatialSearchStrategy_GEOHASH_PREFIX_TREE
	res.SpatialOptionsStrings["WKT"] = spatialOptions
	return res
}

func TestSimonBartlett(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches the order of Java tests
	simonBartlett_circlesShouldNotIntersect(t, driver)
	simonBartlett_lineStringsShouldIntersect(t, driver)
}
