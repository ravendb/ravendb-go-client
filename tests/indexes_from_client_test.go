package tests

import (
	"runtime"
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func NewUsersIndex() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("UsersIndex")
	res.Map = "from user in docs.users select new { user.name }"
	return res
}

func NewUsers_ByName() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("NewUsers_ByName")
	res.Map = "from u in docs.Users select new { u.name }"

	res.Index("name", ravendb.FieldIndexing_SEARCH)

	res.IndexSuggestions.Add("name")

	res.Store("name", ravendb.FieldStorage_YES)

	return res
}

func Posts_ByTitleAndDesc() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("Posts_ByTitleAndDesc")
	res.Map = "from p in docs.Posts select new { p.title, p.desc }"
	res.Index("title", ravendb.FieldIndexing_SEARCH)
	res.Store("title", ravendb.FieldStorage_YES)
	res.Analyze("title", "Lucene.Net.Analysis.SimpleAnalyzer")

	res.Index("desc", ravendb.FieldIndexing_SEARCH)
	res.Store("desc", ravendb.FieldStorage_YES)
	res.Analyze("desc", "Lucene.Net.Analysis.SimpleAnalyzer")

	return res
}

func indexesFromClientTest_canReset(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
	err = store.ExecuteIndex(userIndex)
	assert.NoError(t, err)

	err = gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	command := ravendb.NewGetStatisticsCommand()
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)
	statistics := command.Result
	firstIndexingTime := statistics.GetIndexes()[0].GetLastIndexingTime()

	indexName := NewUsersIndex().GetIndexName()
	// now reset index
	time.Sleep(time.Millisecond * 2)
	{
		op := ravendb.NewResetIndexOperation(indexName)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	err = gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	command = ravendb.NewGetStatisticsCommand()
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)
	statistics = command.Result
	secondIndexingTime := statistics.GetLastIndexingTime()
	assert.True(t, secondIndexingTime.Sub(firstIndexingTime) > 0)
}

func indexesFromClientTest_canExecuteManyIndexes(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	indexes := []*ravendb.AbstractIndexCreationTask{NewUsersIndex()}
	err = store.ExecuteIndexes(indexes)
	assert.NoError(t, err)

	indexNamesOperation := ravendb.NewGetIndexNamesOperation(0, 10)
	err = store.Maintenance().Send(indexNamesOperation)
	assert.NoError(t, err)
	indexNames := indexNamesOperation.Command.Result
	assert.Equal(t, len(indexNames), 1)
}

func indexesFromClientTest_canDelete(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	userIndex := NewUsersIndex()
	err = store.ExecuteIndex(userIndex)
	assert.NoError(t, err)

	op := ravendb.NewDeleteIndexOperation(NewUsersIndex().GetIndexName())
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)

	command := ravendb.NewGetStatisticsCommand()
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)
	statistics := command.Result
	assert.Equal(t, len(statistics.GetIndexes()), 0)
}

func indexesFromClientTest_canStopAndStart(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	err = NewUsers_ByName().Execute(store)
	assert.NoError(t, err)

	{
		op := ravendb.NewGetIndexingStatusOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		status := op.Command.Result

		assert.Equal(t, status.GetStatus(), ravendb.IndexRunningStatus_RUNNING)
		assert.Equal(t, len(status.GetIndexes()), 1)
		assert.Equal(t, status.GetIndexes()[0].GetStatus(), ravendb.IndexRunningStatus_RUNNING)
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
			assert.Equal(t, status.GetStatus(), ravendb.IndexRunningStatus_PAUSED)
			assert.Equal(t, status.GetIndexes()[0].GetStatus(), ravendb.IndexRunningStatus_PAUSED)
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
			indexName = status.GetIndexes()[0].GetName()

			assert.Equal(t, status.GetStatus(), ravendb.IndexRunningStatus_RUNNING)
			assert.Equal(t, len(status.GetIndexes()), 1)
			assert.Equal(t, status.GetIndexes()[0].GetStatus(), ravendb.IndexRunningStatus_RUNNING)
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
			assert.Equal(t, status.GetStatus(), ravendb.IndexRunningStatus_RUNNING)
			assert.Equal(t, len(status.GetIndexes()), 1)
			assert.Equal(t, status.GetIndexes()[0].GetStatus(), ravendb.IndexRunningStatus_PAUSED)
		}
	}
}

func indexesFromClientTest_setLockModeAndSetPriority(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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

	{
		session := openSessionMust(t, store)
		q := session.Query(ravendb.GetTypeOf(&User{}))
		q = q.WaitForNonStaleResults(0)
		// TODO: should this be Name (name of the struct field) and we would
		// convert that to json tag (if necessary) internally?
		q = q.WhereEquals("name", "Arek")
		users, err := q.ToList()
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
		assert.Equal(t, stats.GetLockMode(), ravendb.IndexLockMode_UNLOCK)
		assert.Equal(t, stats.GetPriority(), ravendb.IndexPriority_NORMAL)
	}

	{
		op := ravendb.NewSetIndexesLockOperation(index.Name, ravendb.IndexLockMode_LOCKED_IGNORE)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewSetIndexesPriorityOperation(index.Name, ravendb.IndexPriority_LOW)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}
	{
		op := ravendb.NewGetIndexStatisticsOperation(index.Name)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		stats := op.Command.Result
		assert.Equal(t, stats.GetLockMode(), ravendb.IndexLockMode_LOCKED_IGNORE)
		assert.Equal(t, stats.GetPriority(), ravendb.IndexPriority_LOW)
	}
}

func indexesFromClientTest_getTerms(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
		q := session.Query(ravendb.GetTypeOf(&User{}))
		q = q.WaitForNonStaleResults(0)
		q = q.Statistics(&stats)
		q = q.WhereEquals("name", "Arek")
		_, err := q.ToList()
		assert.NoError(t, err)

		indexName = stats.GetIndexName()

		session.Close()
	}

	op := ravendb.NewGetTermsOperationWithPageSize(indexName, "name", "", 128)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)
	terms := op.Command.Result
	assert.Equal(t, len(terms), 2)
	assert.True(t, ravendb.StringArrayContains(terms, "fitzchak"))
	assert.True(t, ravendb.StringArrayContains(terms, "arek"))
}

func indexesFromClientTest_getIndexNames(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
		q := session.Query(ravendb.GetTypeOf(&User{}))
		q = q.WaitForNonStaleResults(0)
		q = q.Statistics(&stats)
		q = q.WhereEquals("name", "Arek")
		_, err := q.ToList()
		assert.NoError(t, err)

		indexName = stats.GetIndexName()

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		op := ravendb.NewGetIndexNamesOperation(0, 10)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)

		indexNames := op.Command.Result

		assert.Equal(t, len(indexNames), 1)
		assert.True(t, ravendb.StringArrayContains(indexNames, indexName))
		session.Close()
	}
}

func indexesFromClientTest_canExplain(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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

		var statsRef *ravendb.QueryStatistics
		q := session.Query(ravendb.GetTypeOf(&User{}))
		q = q.Statistics(&statsRef)
		q = q.WhereEquals("name", "Arek")
		_, err = q.ToList()
		assert.NoError(t, err)

		q = session.Query(ravendb.GetTypeOf(&User{}))
		q = q.Statistics(&statsRef)
		q = q.WhereGreaterThan("age", 10)
		_, err = q.ToList()
		assert.NoError(t, err)

		session.Close()
	}

	indexQuery := ravendb.NewIndexQuery("from users")
	command := ravendb.NewExplainQueryCommand(store.GetConventions(), indexQuery)
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)

	explanations := command.Result

	assert.Equal(t, len(explanations), 1)
	explanation := explanations[0]
	assert.NotEmpty(t, explanation.GetIndex())
	assert.NotEmpty(t, explanation.GetReason())
}

func indexesFromClientTest_moreLikeThis(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		post1 := NewPost()
		post1.setId("posts/1")
		post1.setTitle("doduck")
		post1.setDesc("prototype")
		err = session.Store(post1)
		assert.NoError(t, err)

		post2 := NewPost()
		post2.setId("posts/2")
		post2.setTitle("doduck")
		post2.setDesc("prototype your idea")
		err = session.Store(post2)
		assert.NoError(t, err)

		post3 := NewPost()
		post3.setId("posts/3")
		post3.setTitle("doduck")
		post3.setDesc("love programming")
		err = session.Store(post3)
		assert.NoError(t, err)

		post4 := NewPost()
		post4.setId("posts/4")
		post4.setTitle("We do")
		post4.setDesc("prototype")
		err = session.Store(post4)
		assert.NoError(t, err)

		post5 := NewPost()
		post5.setId("posts/5")
		post5.setTitle("We love")
		post5.setDesc("challange")
		err = session.Store(post5)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	err = Posts_ByTitleAndDesc().Execute(store)
	assert.NoError(t, err)

	err = gRavenTestDriver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		options := ravendb.NewMoreLikeThisOptions()
		options.SetMinimumDocumentFrequency(1)
		options.SetMinimumTermFrequency(0)

		q := session.QueryInIndex(ravendb.GetTypeOf(&Post{}), Posts_ByTitleAndDesc())

		// TODO: finish this
		/*
			   session.query(Post.class, Posts_ByTitleAndDesc.class)

			   fn1 := func(x *IFilterDocumentQueryBase) IMoreLikeThisOperations {
				   return x.whereEquals("id()", "posts/1")
			   }

			   fn2 := func(f) *DocumentQuery {
				   return f.usingDocumentWithBuilder(fn1).WithOptions(options)
			   }

			   q = q.moreLikeThisWithBuilder(fn2)
		*/

		/*
		   List<Post> list = session.query(Post.class, Posts_ByTitleAndDesc.class)
		           .moreLikeThis(f -> f.usingDocument(x -> x.whereEquals("id()", "posts/1")).WithOptions(options))
		           .toList();
		*/

		_, err = q.ToList()
		assert.NoError(t, err)

		/*
		   assertThat(list)
		           .hasSize(3);

		   assertThat(list.get(0).getTitle())
		           .isEqualTo("doduck");
		   assertThat(list.get(0).getDesc())
		           .isEqualTo("prototype your idea");

		   assertThat(list.get(1).getTitle())
		           .isEqualTo("doduck");
		   assertThat(list.get(1).getDesc())
		           .isEqualTo("love programming");

		   assertThat(list.get(2).getTitle())
		           .isEqualTo("We do");
		   assertThat(list.get(2).getDesc())
		           .isEqualTo("prototype");
		*/

		session.Close()
	}

}

func TestIndexesFromClient(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// order matches Java tests
	indexesFromClientTest_canExecuteManyIndexes(t)
	indexesFromClientTest_canDelete(t)
	indexesFromClientTest_canReset(t)
	indexesFromClientTest_getIndexNames(t)
	indexesFromClientTest_canStopAndStart(t)
	indexesFromClientTest_canExplain(t)
	indexesFromClientTest_moreLikeThis(t)
	// TODO: this works on Mac but fails on Travis CI/Linux
	// https://travis-ci.org/kjk/ravendb-go-client/builds/410576496
	// also sometimes fails on macbook pro
	if ravendb.EnableFailingTests && runtime.GOOS != "linux" {
		indexesFromClientTest_setLockModeAndSetPriority(t)
	}
	indexesFromClientTest_getTerms(t)
}
