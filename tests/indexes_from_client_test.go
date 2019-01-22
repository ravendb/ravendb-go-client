package tests

import (
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func indexesFromClientTestCanCreateIndexesUsingIndexCreation(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	indexes := []*ravendb.AbstractIndexCreationTask{NewUsers_ByName()}
	err = ravendb.IndexCreationCreateIndexes(indexes, store, nil)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		user1 := &User{}
		user1.setName("Marcin")
		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.QueryInIndex(indexes[0])
		err := q.GetResults(&users)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(users))

		session.Close()
	}
}

func indexesFromClientTestCanReset(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user1 := &User{}
		user1.setName("Marcin")
		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	userIndex := NewUsersIndex()
	err = store.ExecuteIndex(userIndex, "")
	assert.NoError(t, err)

	err = driver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	command := ravendb.NewGetStatisticsCommand()
	err = store.GetRequestExecutor("").ExecuteCommand(command)
	assert.NoError(t, err)
	statistics := command.Result
	firstIndexingTime := statistics.Indexes[0].GetLastIndexingTime()

	indexName := NewUsersIndex().GetIndexName()
	// now reset index
	time.Sleep(time.Millisecond * 2)
	{
		op := ravendb.NewResetIndexOperation(indexName)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	err = driver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	command = ravendb.NewGetStatisticsCommand()
	err = store.GetRequestExecutor("").ExecuteCommand(command)
	assert.NoError(t, err)
	statistics = command.Result
	secondIndexingTime := statistics.GetLastIndexingTime()
	assert.True(t, secondIndexingTime.Sub(firstIndexingTime) > 0)
}

func indexesFromClientTestCanExecuteManyIndexes(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	indexes := []*ravendb.AbstractIndexCreationTask{NewUsersIndex()}
	err = store.ExecuteIndexes(indexes, "")
	assert.NoError(t, err)

	indexNamesOperation := ravendb.NewGetIndexNamesOperation(0, 10)
	err = store.Maintenance().Send(indexNamesOperation)
	assert.NoError(t, err)
	indexNames := indexNamesOperation.Command.Result
	assert.Equal(t, len(indexNames), 1)
}

func NewUsersIndex() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("UsersIndex")
	res.Map = "from user in docs.users select new { user.name }"
	return res
}

func indexesFromClientTestCanDelete(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	userIndex := NewUsersIndex()
	err = store.ExecuteIndex(userIndex, "")
	assert.NoError(t, err)

	op := ravendb.NewDeleteIndexOperation(NewUsersIndex().GetIndexName())
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)

	command := ravendb.NewGetStatisticsCommand()
	err = store.GetRequestExecutor("").ExecuteCommand(command)
	assert.NoError(t, err)
	statistics := command.Result
	assert.Equal(t, len(statistics.Indexes), 0)
}

func indexesFromClientTestCanStopAndStart(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	err = NewUsers_ByName().Execute(store, nil, "")
	assert.NoError(t, err)

	{
		op := ravendb.NewGetIndexingStatusOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		status := op.Command.Result

		assert.Equal(t, status.Status, ravendb.IndexRunningStatusRunning)
		assert.Equal(t, len(status.Indexes), 1)
		assert.Equal(t, status.Indexes[0].Status, ravendb.IndexRunningStatusRunning)
	}

	{
		op := ravendb.NewStopIndexingOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)

		{
			op := ravendb.NewGetIndexingStatusOperation()
			err = store.Maintenance().Send(op)
			assert.NoError(t, err)
			status := op.Command.Result
			assert.Equal(t, status.Status, ravendb.IndexRunningStatusPaused)
			assert.Equal(t, status.Indexes[0].Status, ravendb.IndexRunningStatusPaused)
		}
	}

	indexName := ""
	{
		op := ravendb.NewStartIndexingOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		{
			op := ravendb.NewGetIndexingStatusOperation()
			err = store.Maintenance().Send(op)
			assert.NoError(t, err)
			status := op.Command.Result
			indexName = status.Indexes[0].Name

			assert.Equal(t, status.Status, ravendb.IndexRunningStatusRunning)
			assert.Equal(t, len(status.Indexes), 1)
			assert.Equal(t, status.Indexes[0].Status, ravendb.IndexRunningStatusRunning)
		}

	}

	{
		op := ravendb.NewStopIndexOperation(indexName)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		{
			op := ravendb.NewGetIndexingStatusOperation()
			err = store.Maintenance().Send(op)
			assert.NoError(t, err)
			status := op.Command.Result
			assert.Equal(t, status.Status, ravendb.IndexRunningStatusRunning)
			assert.Equal(t, len(status.Indexes), 1)
			assert.Equal(t, status.Indexes[0].Status, ravendb.IndexRunningStatusPaused)
		}
	}
}

func NewUsers_ByName() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("NewUsers_ByName")
	res.Map = "from u in docs.Users select new { u.name }"

	res.Index("name", ravendb.FieldIndexingSearch)

	res.IndexSuggestions = append(res.IndexSuggestions, "name")

	res.Store("name", ravendb.FieldStorageYes)

	return res
}

func indexesFromClientTestSetLockModeAndSetPriority(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	usersByName := NewUsers_ByName()
	err = store.ExecuteIndex(usersByName, "")
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		user1 := &User{}
		user1.setName("Fitzchak")
		err = session.Store(user1)
		assert.NoError(t, err)

		user2 := &User{}
		user2.setName("Arek")
		err = session.Store(user2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.QueryInIndex(usersByName)
		q = q.WaitForNonStaleResults(0)
		q = q.WhereEquals("name", "Arek")
		err := q.GetResults(&users)
		assert.NoError(t, err)
		assert.Equal(t, len(users), 1)
	}

	op := ravendb.NewGetIndexesOperation(0, 128)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)
	indexes := op.Command.Result
	assert.Equal(t, len(indexes), 1)

	index := indexes[0]

	{
		op := ravendb.NewGetIndexStatisticsOperation(index.Name)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		stats := op.Command.Result
		assert.Equal(t, stats.LockMode, ravendb.IndexLockModeUnlock)
		assert.Equal(t, stats.Priority, ravendb.IndexPriorityNormal)
	}

	{
		op := ravendb.NewSetIndexesLockOperation(index.Name, ravendb.IndexLockModeLockedIgnore)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewSetIndexesPriorityOperation(index.Name, ravendb.IndexPriorityLow)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}
	{
		op := ravendb.NewGetIndexStatisticsOperation(index.Name)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		stats := op.Command.Result
		assert.Equal(t, stats.LockMode, ravendb.IndexLockModeLockedIgnore)
		assert.Equal(t, stats.Priority, ravendb.IndexPriorityLow)
	}
}

func indexesFromClientTestGetTerms(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user1 := &User{}
		user1.setName("Fitzchak")
		err = session.Store(user1)
		assert.NoError(t, err)

		user2 := &User{}
		user2.setName("Arek")
		err = session.Store(user2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	var indexName string

	{
		session := openSessionMust(t, store)

		var stats *ravendb.QueryStatistics
		var notUsed []*User
		q := session.Query()
		q = q.WaitForNonStaleResults(0)
		q = q.Statistics(&stats)
		q = q.WhereEquals("name", "Arek")
		err := q.GetResults(&notUsed)
		assert.NoError(t, err)

		indexName = stats.IndexName

		session.Close()
	}

	op := ravendb.NewGetTermsOperation(indexName, "name", "", 128)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)
	terms := op.Command.Result
	assert.Equal(t, len(terms), 2)
	assert.True(t, stringArrayContains(terms, "fitzchak"))
	assert.True(t, stringArrayContains(terms, "arek"))
}

func indexesFromClientTestGetIndexNames(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user1 := &User{}
		user1.setName("Fitzchak")
		err = session.Store(user1)
		assert.NoError(t, err)

		user2 := &User{}
		user2.setName("Arek")
		err = session.Store(user2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	var indexName string

	{
		session := openSessionMust(t, store)

		var notUsed []*User
		var stats *ravendb.QueryStatistics
		q := session.Query()
		q = q.WaitForNonStaleResults(0)
		q = q.Statistics(&stats)
		q = q.WhereEquals("name", "Arek")
		err := q.GetResults(&notUsed)
		assert.NoError(t, err)

		indexName = stats.IndexName

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		op := ravendb.NewGetIndexNamesOperation(0, 10)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)

		indexNames := op.Command.Result

		assert.Equal(t, len(indexNames), 1)
		assert.True(t, stringArrayContains(indexNames, indexName))
		session.Close()
	}
}

func indexesFromClientTestCanExplain(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	user1 := &User{}
	user1.setName("Fitzchak")

	user2 := &User{}
	user2.setName("Arek")

	{
		session := openSessionMust(t, store)

		err = session.Store(user1)
		assert.NoError(t, err)
		err = session.Store(user2)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var notUsed []*User
		var statsRef *ravendb.QueryStatistics
		q := session.Query()
		q = q.Statistics(&statsRef)
		q = q.WhereEquals("name", "Arek")
		err = q.GetResults(&notUsed)
		assert.NoError(t, err)

		q = session.Query()
		q = q.Statistics(&statsRef)
		q = q.WhereGreaterThan("age", 10)
		err = q.GetResults(&notUsed)
		assert.NoError(t, err)

		session.Close()
	}

	indexQuery := ravendb.NewIndexQuery("from users")
	command := ravendb.NewExplainQueryCommand(store.GetConventions(), indexQuery)
	err = store.GetRequestExecutor("").ExecuteCommand(command)
	assert.NoError(t, err)

	explanations := command.Result

	assert.Equal(t, len(explanations), 1)
	explanation := explanations[0]
	assert.NotEmpty(t, explanation.Index)
	assert.NotEmpty(t, explanation.Reason)
}

func indexesFromClientTestMoreLikeThis(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		post1 := &Post{
			ID:    "posts/1",
			Title: "doduck",
			Desc:  "prototype",
		}
		err = session.Store(post1)
		assert.NoError(t, err)

		post2 := &Post{
			ID:    "posts/2",
			Title: "doduck",
			Desc:  "prototype your idea",
		}
		err = session.Store(post2)
		assert.NoError(t, err)

		post3 := &Post{
			ID:    "posts/3",
			Title: "doduck",
			Desc:  "love programming",
		}
		err = session.Store(post3)
		assert.NoError(t, err)

		post4 := &Post{
			ID:    "posts/4",
			Title: "We do",
			Desc:  "prototype",
		}
		err = session.Store(post4)
		assert.NoError(t, err)

		post5 := &Post{
			ID:    "posts/5",
			Title: "We love",
			Desc:  "challenge",
		}
		err = session.Store(post5)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	err = PostsByTitleAndDesc().Execute(store, nil, "")
	assert.NoError(t, err)

	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		options := ravendb.NewMoreLikeThisOptions()
		options.SetMinimumDocumentFrequency(1)
		options.SetMinimumTermFrequency(0)

		var list []*Post
		q := session.QueryInIndex(PostsByTitleAndDesc())

		fn1 := func(x *ravendb.DocumentQuery) {
			x.WhereEquals("id()", "posts/1")
		}

		fn2 := func(f ravendb.IMoreLikeThisBuilderForDocumentQuery) {
			f.UsingDocumentWithBuilder(fn1).WithOptions(options)
		}

		q = q.MoreLikeThisWithBuilder(fn2)

		err := q.GetResults(&list)
		assert.NoError(t, err)

		assert.Equal(t, len(list), 3)

		var post *Post

		post = list[0]
		assert.Equal(t, post.Title, "doduck")
		assert.Equal(t, post.Desc, "prototype your idea")

		post = list[1]
		assert.Equal(t, post.Title, "doduck")
		assert.Equal(t, post.Desc, "love programming")

		post = list[2]
		assert.Equal(t, post.Title, "We do")
		assert.Equal(t, post.Desc, "prototype")

		session.Close()
	}
}

func PostsByTitleAndDesc() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("Posts_ByTitleAndDesc")
	res.Map = "from p in docs.Posts select new { p.title, p.desc }"
	res.Index("title", ravendb.FieldIndexingSearch)
	res.Store("title", ravendb.FieldStorageYes)
	res.Analyze("title", "Lucene.Net.Analysis.SimpleAnalyzer")

	res.Index("desc", ravendb.FieldIndexingSearch)
	res.Store("desc", ravendb.FieldStorageYes)
	res.Analyze("desc", "Lucene.Net.Analysis.SimpleAnalyzer")

	return res
}

func TestIndexesFromClient(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// order matches Java tests
	indexesFromClientTestCanExecuteManyIndexes(t, driver)
	indexesFromClientTestCanDelete(t, driver)
	indexesFromClientTestCanReset(t, driver)
	indexesFromClientTestGetIndexNames(t, driver)
	indexesFromClientTestCanStopAndStart(t, driver)
	indexesFromClientTestCanExplain(t, driver)
	indexesFromClientTestMoreLikeThis(t, driver)
	indexesFromClientTestSetLockModeAndSetPriority(t, driver)
	indexesFromClientTestGetTerms(t, driver)

	// TODO: order doesn't match Java
	indexesFromClientTestCanCreateIndexesUsingIndexCreation(t, driver)
}
