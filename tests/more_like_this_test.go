package tests

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func getLorem(numWords int) string {
	theLorem := "Morbi nec purus eu libero interdum laoreet Nam metus quam posuere in elementum eget egestas eget justo Aenean orci ligula ullamcorper nec convallis non placerat nec lectus Quisque convallis porta suscipit Aliquam sollicitudin ligula sit amet libero cursus egestas Maecenas nec mauris neque at faucibus justo Fusce ut orci neque Nunc sodales pulvinar lobortis Praesent dui tellus fermentum sed faucibus nec faucibus non nibh Vestibulum adipiscing porta purus ut varius mi pulvinar eu Nam sagittis sodales hendrerit Vestibulum et tincidunt urna Fusce lacinia nisl at luctus lobortis lacus quam rhoncus risus a posuere nulla lorem at nisi Sed non erat nisl Cras in augue velit a mattis ante Etiam lorem dui elementum eget facilisis vitae viverra sit amet tortor Suspendisse potenti Nunc egestas accumsan justo viverra viverra Sed faucibus ullamcorper mauris ut pharetra ligula ornare eget Donec suscipit luctus rhoncus Pellentesque eget justo ac nunc tempus consequat Nullam fringilla egestas leo Praesent condimentum laoreet magna vitae luctus sem cursus sed Mauris massa purus suscipit ac malesuada a accumsan non neque Proin et libero vitae quam ultricies rhoncus Praesent urna neque molestie et suscipit vestibulum iaculis ac nulla Integer porta nulla vel leo ullamcorper eu rhoncus dui semper Donec dictum dui"

	loremArray := strings.Split(theLorem, " ")

	output := ""
	maxN := len(loremArray) - 1

	for i := 0; i < numWords; i++ {
		idx := rand.Intn(maxN)
		s := loremArray[idx]
		output += s
		output += " "
	}
	return output
}

func getDataList() []*Data {
	var items []*Data
	items = append(items, NewData("This is a test. Isn't it great? I hope I pass my test!"))
	items = append(items, NewData("I have a test tomorrow. I hate having a test"))
	items = append(items, NewData("Cake is great."))
	items = append(items, NewData("This document has the word test only once"))
	items = append(items, NewData("test"))
	items = append(items, NewData("test"))
	items = append(items, NewData("test"))
	items = append(items, NewData("test"))
	return items
}

func moreLikeThis_canGetResultsUsingTermVectors(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	var id string
	dataIndex := NewDataIndex2(true, false)
	{
		session := openSessionMust(t, store)
		dataIndex.Execute(store)
		list := getDataList()
		for _, el := range list {
			err = session.Store(el)
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		id = session.Advanced().GetDocumentID(list[0])
		gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	}

	moreLikeThis_assertMoreLikeThisHasMatchesFor(t, reflect.TypeOf(&Data{}), dataIndex, store, id)
}

func moreLikeThis_canGetResultsUsingTermVectorsLazy(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	var id string
	dataIndex := NewDataIndex2(true, false)

	{
		session := openSessionMust(t, store)
		dataIndex.Execute(store)
		list := getDataList()
		for _, el := range list {
			err = session.Store(el)
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		id = session.Advanced().GetDocumentID(list[0])
		gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	}
	{
		session := openSessionMust(t, store)
		options := ravendb.NewMoreLikeThisOptions()
		options.Fields = []string{"body"}

		query := session.QueryInIndexOld(reflect.TypeOf(&Data{}), dataIndex)
		builder := func(f ravendb.IMoreLikeThisBuilderForDocumentQuery) {
			builder := func(b *ravendb.IFilterDocumentQueryBase) {
				b.WhereEquals("id()", id)
			}
			ops := f.UsingDocumentWithBuilder(builder)
			ops.WithOptions(options)
		}
		query = query.MoreLikeThisWithBuilder(builder)
		lazyLst := query.Lazily()
		list, err := lazyLst.GetValue()
		assert.NoError(t, err)
		v := list.([]*Data)
		assert.NotEmpty(t, v)
		// TODO: more precise check that returned the right values
	}
}

func moreLikeThis_canGetResultsUsingTermVectorsWithDocumentQuery(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	var id string
	dataIndex := NewDataIndex2(true, false)

	{
		session := openSessionMust(t, store)
		dataIndex.Execute(store)
		list := getDataList()
		for _, el := range list {
			err = session.Store(el)
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		id = session.Advanced().GetDocumentID(list[0])
		gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	}
	{
		session := openSessionMust(t, store)
		options := ravendb.NewMoreLikeThisOptions()
		options.Fields = []string{"body"}
		query := session.QueryInIndexOld(reflect.TypeOf(&Data{}), dataIndex)
		builder := func(f ravendb.IMoreLikeThisBuilderForDocumentQuery) {
			builder := func(b *ravendb.IFilterDocumentQueryBase) {
				b.WhereEquals("id()", id)
			}
			ops := f.UsingDocumentWithBuilder(builder)
			ops.WithOptions(options)
		}
		query = query.MoreLikeThisWithBuilder(builder)
		var list []*Data
		err = query.ToList(&list)
		assert.NoError(t, err)
		assert.NotEmpty(t, list)
		assert.Equal(t, len(list), 7)
		// TODO: better check if returned the right result
	}
}

func moreLikeThis_canGetResultsUsingStorage(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	var id string
	dataIndex := NewDataIndex2(false, true)

	{
		session := openSessionMust(t, store)
		dataIndex.Execute(store)
		list := getDataList()
		for _, el := range list {
			err = session.Store(el)
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		id = session.Advanced().GetDocumentID(list[0])
		gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	}

	moreLikeThis_assertMoreLikeThisHasMatchesFor(t, reflect.TypeOf(&Data{}), dataIndex, store, id)
}

func moreLikeThis_canGetResultsUsingTermVectorsAndStorage(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	var id string
	dataIndex := NewDataIndex2(true, true)

	{
		session := openSessionMust(t, store)
		dataIndex.Execute(store)
		list := getDataList()
		for _, el := range list {
			err = session.Store(el)
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		id = session.Advanced().GetDocumentID(list[0])
		gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	}
	moreLikeThis_assertMoreLikeThisHasMatchesFor(t, reflect.TypeOf(&Data{}), dataIndex, store, id)
}

func moreLikeThis_test_With_Lots_Of_Random_Data(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	key := "data/1-A" // Note: in Java it's datas/ because of bad pluralization of data
	dataIndex := NewDataIndex()

	{
		session := openSessionMust(t, store)
		dataIndex.Execute(store)
		for i := 0; i < 100; i++ {
			data := &Data{}
			data.Body = getLorem(200)
			err = session.Store(data)
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)

	}
	moreLikeThis_assertMoreLikeThisHasMatchesFor(t, reflect.TypeOf(&Data{}), dataIndex, store, key)
}

func moreLikeThis_do_Not_Pass_FieldNames(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	key := "data/1-A" // Note: in Java it's datas/ because of bad pluralization of data
	dataIndex := NewDataIndex()

	{
		session := openSessionMust(t, store)
		dataIndex.Execute(store)
		for i := 0; i < 10; i++ {
			data := &Data{}
			data.Body = fmt.Sprintf("Body%d", i)
			data.WhitespaceAnalyzerField = "test test"
			err = session.Store(data)
			assert.NoError(t, err)

		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	}

	{
		session := openSessionMust(t, store)
		query := session.QueryInIndexOld(reflect.TypeOf(&Data{}), dataIndex)
		builder := func(f ravendb.IMoreLikeThisBuilderForDocumentQuery) {
			builder := func(b *ravendb.IFilterDocumentQueryBase) {
				b.WhereEquals("id()", key)
			}
			f.UsingDocumentWithBuilder(builder)
		}
		query = query.MoreLikeThisWithBuilder(builder)
		var list []*Data
		err = query.ToList(&list)
		assert.NoError(t, err)
		assert.Equal(t, 9, len(list)) // TODO: should this be 10? 1?
	}
}

func moreLikeThis_each_Field_Should_Use_Correct_Analyzer(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	key1 := "data/1-A" // Note: in Java it's datas/ because of bad pluralization of data
	dataIndex := NewDataIndex()

	{
		session := openSessionMust(t, store)
		dataIndex.Execute(store)
		for i := 0; i < 10; i++ {
			data := &Data{}
			data.WhitespaceAnalyzerField = "bob@hotmail.com hotmail"
			err = session.Store(data)
			assert.NoError(t, err)

		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	}

	{
		session := openSessionMust(t, store)
		options := ravendb.NewMoreLikeThisOptions()
		v1 := 2
		options.MinimumTermFrequency = &v1
		v2 := 5
		options.MinimumDocumentFrequency = &v2

		query := session.QueryInIndexOld(reflect.TypeOf(&Data{}), dataIndex)
		builder := func(f ravendb.IMoreLikeThisBuilderForDocumentQuery) {
			builder := func(b *ravendb.IFilterDocumentQueryBase) {
				b.WhereEquals("id()", key1)
			}
			o := f.UsingDocumentWithBuilder(builder)
			o.WithOptions(options)
		}
		query = query.MoreLikeThisWithBuilder(builder)
		var list []*Data
		err = query.ToList(&list)
		assert.NoError(t, err)
		assert.Empty(t, list)
	}

	key2 := "data/11-A" // Note: in Java it's datas/ because of bad pluralization of data

	{
		session := openSessionMust(t, store)
		dataIndex.Execute(store)
		for i := 0; i < 10; i++ {
			data := &Data{}
			data.WhitespaceAnalyzerField = "bob@hotmail.com bob@hotmail.com"
			err = session.Store(data)
			assert.NoError(t, err)

		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	}

	{
		session := openSessionMust(t, store)

		query := session.QueryInIndexOld(reflect.TypeOf(&Data{}), dataIndex)
		builder := func(f ravendb.IMoreLikeThisBuilderForDocumentQuery) {
			builder := func(b *ravendb.IFilterDocumentQueryBase) {
				b.WhereEquals("id()", key2)
			}
			f.UsingDocumentWithBuilder(builder)
		}
		query = query.MoreLikeThisWithBuilder(builder)
		var list []*Data
		err = query.ToList(&list)
		assert.NoError(t, err)
		assert.NotEmpty(t, list)
	}
}

func moreLikeThis_can_Use_Min_Doc_Freq_Param(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	key := "data/1-A" // Note: in Java it's datas/ because of bad pluralization of data
	dataIndex := NewDataIndex()

	{
		session := openSessionMust(t, store)
		dataIndex.Execute(store)
		d := []string{
			"This is a test. Isn't it great? I hope I pass my test!",
			"I have a test tomorrow. I hate having a test",
			"Cake is great.",
			"This document has the word test only once",
		}
		for _, text := range d {
			data := &Data{}
			data.Body = text
			err = session.Store(data)
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	}

	{
		session := openSessionMust(t, store)
		options := ravendb.NewMoreLikeThisOptions()
		options.Fields = []string{"body"}
		options.SetMinimumDocumentFrequency(2)

		query := session.QueryInIndexOld(reflect.TypeOf(&Data{}), dataIndex)
		builder := func(f ravendb.IMoreLikeThisBuilderForDocumentQuery) {
			builder := func(b *ravendb.IFilterDocumentQueryBase) {
				b.WhereEquals("id()", key)
			}
			o := f.UsingDocumentWithBuilder(builder)
			o.WithOptions(options)
		}
		query = query.MoreLikeThisWithBuilder(builder)
		var list []*Data
		err = query.ToList(&list)
		assert.NoError(t, err)
		assert.NotEmpty(t, list)
	}
}

func moreLikeThis_can_Use_Boost_Param(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	key := "data/1-A" // Note: in Java it's datas/ because of bad pluralization of data
	dataIndex := NewDataIndex()

	{
		session := openSessionMust(t, store)
		dataIndex.Execute(store)
		d := []string{
			"This is a test. it is a great test. I hope I pass my great test!",
			"Cake is great.",
			"I have a test tomorrow.",
		}
		for _, text := range d {
			data := &Data{
				Body: text,
			}
			err = session.Store(data)
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	}

	{
		session := openSessionMust(t, store)
		options := ravendb.NewMoreLikeThisOptions()
		options.Fields = []string{"body"}
		options.SetMinimumWordLength(3)
		options.SetMinimumDocumentFrequency(1)
		options.SetMinimumTermFrequency(2)
		options.SetBoost(true)

		query := session.QueryInIndexOld(reflect.TypeOf(&Data{}), dataIndex)
		builder := func(f ravendb.IMoreLikeThisBuilderForDocumentQuery) {
			builder := func(b *ravendb.IFilterDocumentQueryBase) {
				b.WhereEquals("id()", key)
			}
			o := f.UsingDocumentWithBuilder(builder)
			o.WithOptions(options)
		}
		query = query.MoreLikeThisWithBuilder(builder)
		var list []*Data
		err = query.ToList(&list)
		assert.NoError(t, err)
		assert.NotEmpty(t, list)
		assert.Equal(t, list[0].Body, "I have a test tomorrow.")
	}
}

func moreLikeThis_can_Use_Stop_Words(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	key := "data/1-A" // Note: in Java it's datas/ because of bad pluralization of data
	dataIndex := NewDataIndex()

	{
		session := openSessionMust(t, store)
		dataIndex.Execute(store)
		d := []string{
			"This is a test. Isn't it great? I hope I pass my test!",
			"I should not hit this document. I hope",
			"Cake is great.",
			"This document has the word test only once",
			"test",
			"test",
			"test",
			"test",
		}
		for _, text := range d {
			data := &Data{
				Body: text,
			}
			err = session.Store(data)
			assert.NoError(t, err)
		}

		stopWords := &ravendb.MoreLikeThisStopWords{
			ID:        "Config/Stopwords",
			StopWords: []string{"I", "A", "Be"},
		}
		err = session.Store(stopWords)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)
		gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	}

	{
		session := openSessionMust(t, store)
		options := ravendb.NewMoreLikeThisOptions()
		options.Fields = []string{"body"}
		options.SetMinimumTermFrequency(2)
		options.SetMinimumDocumentFrequency(1)
		options.SetStopWordsDocumentID("Config/Stopwords")

		query := session.QueryInIndexOld(reflect.TypeOf(&Data{}), dataIndex)
		builder := func(f ravendb.IMoreLikeThisBuilderForDocumentQuery) {
			builder := func(b *ravendb.IFilterDocumentQueryBase) {
				b.WhereEquals("id()", key)
			}
			o := f.UsingDocumentWithBuilder(builder)
			o.WithOptions(options)
		}
		query = query.MoreLikeThisWithBuilder(builder)
		var list []*Data
		err = query.ToList(&list)
		assert.NoError(t, err)
		assert.Equal(t, len(list), 5)
	}
}

func moreLikeThis_canMakeDynamicDocumentQueries(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	dataIndex := NewDataIndex()
	{
		session := openSessionMust(t, store)
		dataIndex.Execute(store)
		list := getDataList()
		for _, el := range list {
			err = session.Store(el)
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)

		gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	}

	{
		session := openSessionMust(t, store)
		options := ravendb.NewMoreLikeThisOptions()
		options.Fields = []string{"body"}
		options.SetMinimumTermFrequency(1)
		options.SetMinimumDocumentFrequency(1)

		query := session.QueryInIndexOld(reflect.TypeOf(&Data{}), dataIndex)
		builder := func(f ravendb.IMoreLikeThisBuilderForDocumentQuery) {
			o := f.UsingDocument("{ \"body\": \"A test\" }")
			o.WithOptions(options)
		}
		query = query.MoreLikeThisWithBuilder(builder)
		var list []*Data
		err = query.ToList(&list)
		assert.NoError(t, err)
		assert.Equal(t, len(list), 7)
	}
}

func moreLikeThis_canMakeDynamicDocumentQueriesWithComplexProperties(t *testing.T) {}

func moreLikeThis_assertMoreLikeThisHasMatchesFor(t *testing.T, clazz reflect.Type, index *ravendb.AbstractIndexCreationTask, store *ravendb.IDocumentStore, documentKey string) {
	session := openSessionMust(t, store)

	options := ravendb.NewMoreLikeThisOptions()
	options.Fields = []string{"body"}

	q := session.QueryInIndexOld(clazz, index)
	fn1 := func(b *ravendb.IFilterDocumentQueryBase) {
		b.WhereEquals("id()", documentKey)
	}
	fn2 := func(f ravendb.IMoreLikeThisBuilderForDocumentQuery) {
		f.UsingDocumentWithBuilder(fn1).WithOptions(options)
	}
	q = q.MoreLikeThisWithBuilder(fn2)
	list, err := q.ToListOld()
	assert.NoError(t, err)
	assert.True(t, len(list) > 0)

	session.Close()
}

type Identity struct {
	ID string
}

type Data struct {
	Identity

	Body                    string `json:"body"`
	WhitespaceAnalyzerField string `json:"whitespaceAnalyzerField"`
	PersonID                string `json:"personId"`
}

func NewData(s string) *Data {
	return &Data{
		Body: s,
	}
}

type DataWithIntegerId struct {
	Identity
	Body string `json:"body"`
}

type ComplexData struct {
	ID       string
	Property *ComplexProperty `json:"property"`
}

type ComplexProperty struct {
	Body string `json:"body"`
}

func NewDataIndex() *ravendb.AbstractIndexCreationTask {
	return NewDataIndex2(true, false)
}

func NewDataIndex2(termVector bool, store bool) *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("DataIndex")

	// Note: in Java it's docs.Datas because Inflector.pluralize() doesn't
	// handle 'data' properly and we do
	res.Map = "from doc in docs.Data select new { doc.body, doc.whitespaceAnalyzerField }"

	res.Analyze("body", "Lucene.Net.Analysis.Standard.StandardAnalyzer")
	res.Analyze("whitespaceAnalyzerField", "Lucene.Net.Analysis.WhitespaceAnalyzer")

	if store {
		res.Store("body", ravendb.FieldStorage_YES)
		res.Store("whitespaceAnalyzerField", ravendb.FieldStorage_YES)
	}

	if termVector {
		res.TermVector("body", ravendb.FieldTermVector_YES)
		res.TermVector("whitespaceAnalyzerField", ravendb.FieldTermVector_YES)
	}
	return res
}

func NewComplexDataIndex() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("ComplexDataIndex")
	res.Map = "from doc in docs.ComplexDatas select new  { doc.property, doc.property.body }"
	res.Index("body", ravendb.FieldIndexing_SEARCH)
	return res
}

func TestMoreLikeThis(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// order matches Java tests
	moreLikeThis_can_Use_Stop_Words(t)
	moreLikeThis_can_Use_Boost_Param(t)
	moreLikeThis_canGetResultsUsingTermVectors(t)
	moreLikeThis_can_Use_Min_Doc_Freq_Param(t)
	moreLikeThis_each_Field_Should_Use_Correct_Analyzer(t)
	moreLikeThis_test_With_Lots_Of_Random_Data(t)
	moreLikeThis_canMakeDynamicDocumentQueries(t)
	moreLikeThis_canGetResultsUsingTermVectorsAndStorage(t)
	moreLikeThis_canGetResultsUsingStorage(t)
	moreLikeThis_canMakeDynamicDocumentQueriesWithComplexProperties(t)
	moreLikeThis_do_Not_Pass_FieldNames(t)
	moreLikeThis_canGetResultsUsingTermVectorsLazy(t)
	moreLikeThis_canGetResultsUsingTermVectorsWithDocumentQuery(t)
}
