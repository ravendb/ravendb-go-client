package tests

import (
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func simonBartlettLineStringsShouldIntersect(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewGeoIndex()
	err = store.ExecuteIndex(index, "")
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

	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		q := session.QueryIndex(index.IndexName)
		fn := func(f *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			return f.RelatesToShape("LINESTRING (1 0, 1 1, 1 2)", ravendb.SpatialRelationIntersects)
		}
		q = q.Spatial3("WKT", fn)
		q = q.WaitForNonStaleResults(0)
		count, err := q.Count()
		assert.NoError(t, err)

		assert.Equal(t, count, 1)

		q = session.QueryIndex(index.IndexName)
		q = q.RelatesToShape("WKT", "LINESTRING (1 0, 1 1, 1 2)", ravendb.SpatialRelationIntersects)
		q = q.WaitForNonStaleResults(0)
		count, err = q.Count()
		assert.NoError(t, err)

		assert.Equal(t, count, 1)

		session.Close()
	}
}

func simonBartlettCirclesShouldNotIntersect(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewGeoIndex()
	err = store.ExecuteIndex(index, "")
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
		q := session.QueryIndex(index.IndexName)
		fn := func(f *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			return f.RelatesToShape("CIRCLE(0.000000 3.000000 d=110)", ravendb.SpatialRelationIntersects)
		}

		q = q.Spatial3("WKT", fn)

		q = q.WaitForNonStaleResults(0)
		count, err := q.Count()
		assert.NoError(t, err)

		assert.Equal(t, count, 0)

		q = session.QueryIndex(index.IndexName)
		q = q.RelatesToShape("WKT", "CIRCLE(0.000000 3.000000 d=110)", ravendb.SpatialRelationIntersects)
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

func NewGeoIndex() *ravendb.IndexCreationTask {
	res := ravendb.NewIndexCreationTask("GeoIndex")
	res.Map = "docs.GeoDocuments.Select(doc => new {\n" +
		"    WKT = this.CreateSpatialField(doc.WKT)\n" +
		"})"
	spatialOptions := ravendb.NewSpatialOptions()
	spatialOptions.Strategy = ravendb.SpatialSearchStrategyGeohashPrefixTree
	res.SpatialOptionsStrings["WKT"] = spatialOptions
	return res
}

func TestSimonBartlett(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches the order of Java tests
	simonBartlettCirclesShouldNotIntersect(t, driver)
	simonBartlettLineStringsShouldIntersect(t, driver)
}
