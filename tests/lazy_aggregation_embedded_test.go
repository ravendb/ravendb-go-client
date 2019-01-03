package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func lazyAggregationEmbeddedLazyTest(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
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
		driver.waitForIndexing(store, "", 0)

		q := session.QueryInIndex(index)
		builder := func(f ravendb.IFacetBuilder) {
			f.ByField("AssigneeID").WithDisplayName("AssigneeID")
		}
		query := q.AggregateBy(builder)
		facetValue := map[string]*ravendb.FacetResult{}
		lazyOperation := query.ExecuteLazy(facetValue, nil)
		err = lazyOperation.GetValue()
		assert.NoError(t, err)
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
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	lazyAggregationEmbeddedLazyTest(t, driver)
}
