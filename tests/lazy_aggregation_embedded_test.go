package tests

import (
	"reflect"
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func lazyAggregationEmbeddedLazy_test(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		index := NewTaskIndex()

		session := openSessionMust(t, store)

		task1 := &Task{
			AssigneeID: "users/1",
			ID:         "tasks/1",
		}

		task2 := &Task{
			AssigneeID: "users/1",
			ID:         "tasks/2",
		}

		task3 := &Task{
			AssigneeID: "users/2",
			ID:         "tasks/3",
		}

		err = session.Store(task1)
		assert.NoError(t, err)
		err = session.Store(task2)
		assert.NoError(t, err)
		err = session.Store(task3)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		index.Execute(store)
		gRavenTestDriver.waitForIndexing(store, "", 0)

		q := session.QueryInIndexOld(reflect.TypeOf(&Order{}), index)
		builder := func(f ravendb.IFacetBuilder) {
			f.ByField("AssigneeID").WithDisplayName("AssigneeID")
		}
		query := q.AggregateBy(builder)
		lazyOperation := query.ExecuteLazy(nil)
		facetValueI, err := lazyOperation.GetValue()
		assert.NoError(t, err)
		facetValue := facetValueI.(map[string]*ravendb.FacetResult)
		values := facetValue["AssigneeID"].Values
		userStats := map[string]int{}
		for _, value := range values {
			r := value.Range
			c := value.Count
			userStats[r] = c
		}
		assert.Equal(t, userStats["users/1"], 2)
		assert.Equal(t, userStats["users/2"], 1)

		session.Close()
	}

}

func NewTaskIndex() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("TaskIndex")
	res.Map = "from task in docs.Tasks select new { task.AssigneeID } "
	return res
}

type Task struct {
	ID         string
	AssigneeID string
}

func TestLazyAggregationEmbeddedLazy(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	lazyAggregationEmbeddedLazy_test(t)
}
