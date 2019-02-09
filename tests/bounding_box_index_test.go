package tests

import (
	"reflect"
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func boundingBox_boundingBoxTest(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	polygon := "POLYGON ((0 0, 0 5, 1 5, 1 1, 5 1, 5 5, 6 5, 6 0, 0 0))"
	rectangle1 := "2 2 4 4"
	rectangle2 := "6 6 10 10"
	rectangle3 := "0 0 6 6"

	bboxIndex := NewBBoxIndex()
	bboxIndex.Execute(store, nil, "")
	quadTreeIndex := NewQuadTreeIndex()
	quadTreeIndex.Execute(store, nil, "")

	{
		session := openSessionMust(t, store)

		doc := &SpatialDoc{
			Shape: polygon,
		}
		err = session.Store(doc)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		clazz := reflect.TypeOf(&SpatialDoc{})
		q, err := session.QueryCollectionForType(clazz)
		assert.NoError(t, err)
		result, err := q.Count()
		assert.NoError(t, err)
		assert.Equal(t, result, 1)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q, err := session.QueryIndex(bboxIndex.IndexName)
		assert.NoError(t, err)
		clause := func(x *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			return x.Intersects(rectangle1)
		}
		q.Spatial3("shape", clause)
		result, err := q.Count()
		assert.NoError(t, err)
		assert.Equal(t, result, 1)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q, err := session.QueryIndex(bboxIndex.IndexName)
		assert.NoError(t, err)
		clause := func(x *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			return x.Intersects(rectangle2)
		}
		q.Spatial3("shape", clause)
		result, err := q.Count()
		assert.NoError(t, err)
		assert.Equal(t, result, 0)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q, err := session.QueryIndex(bboxIndex.IndexName)
		assert.NoError(t, err)
		clause := func(x *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			return x.Disjoint(rectangle2)
		}
		q.Spatial3("shape", clause)
		result, err := q.Count()
		assert.NoError(t, err)
		assert.Equal(t, result, 1)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q, err := session.QueryIndex(bboxIndex.IndexName)
		assert.NoError(t, err)
		clause := func(x *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			return x.Within(rectangle3)
		}
		q.Spatial3("shape", clause)
		result, err := q.Count()
		assert.NoError(t, err)
		assert.Equal(t, result, 1)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q, err := session.QueryIndex(quadTreeIndex.IndexName)
		assert.NoError(t, err)
		clause := func(x *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			return x.Intersects(rectangle2)
		}
		q.Spatial3("shape", clause)
		result, err := q.Count()
		assert.NoError(t, err)
		assert.Equal(t, result, 0)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q, err := session.QueryIndex(quadTreeIndex.IndexName)
		assert.NoError(t, err)
		clause := func(x *ravendb.SpatialCriteriaFactory) ravendb.SpatialCriteria {
			return x.Intersects(rectangle1)
		}
		q.Spatial3("shape", clause)
		result, err := q.Count()
		assert.NoError(t, err)
		assert.Equal(t, result, 0)

		session.Close()
	}
}

type SpatialDoc struct {
	ID    string
	Shape string `json:"shape"`
}

func NewBBoxIndex() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("BBoxIndex")
	res.Map = "docs.SpatialDocs.Select(doc => new {\n" +
		"    shape = this.CreateSpatialField(doc.shape)\n" +
		"})"
	indexing := func() *ravendb.SpatialOptions {
		return ravendb.NewGeograpyboundingBoxIndex()
	}
	res.Spatial("shape", indexing)
	return res
}

func NewQuadTreeIndex() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("QuadTreeIndex")

	res.Map = `docs.SpatialDocs.Select(doc => new {
   shape = this.CreateSpatialField(doc.shape)
})`
	indexing := func() *ravendb.SpatialOptions {
		bounds := ravendb.NewSpatialBounds(0, 0, 16, 16)
		return ravendb.NewCartesianQuadPrefixTreeIndex(6, bounds)
	}
	res.Spatial("shape", indexing)
	return res
}

func TestBoundingBox(t *testing.T) {
	// // t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	boundingBox_boundingBoxTest(t, driver)
}
