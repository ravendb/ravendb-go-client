package tests

import (
	"math"
	"sort"
	"strconv"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func collectUserNamesSorted(a []*User) []string {
	var names []string
	for _, user := range a {
		names = append(names, *user.Name)
	}
	sort.Strings(names)
	return names
}

func createRevisions(t *testing.T, store *ravendb.DocumentStore) {
	for i := 0; i < 4; i++ {
		session := openSessionMust(t, store)
		user := &User{}
		user.setName("user" + strconv.Itoa(i+1))
		err := session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}
}

func revisionsTestRevisions(t *testing.T, driver *RavenTestDriver) {

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	_, err = setupRevisions(store, false, 4)
	assert.NoError(t, err)

	createRevisions(t, store)

	{
		session := openSessionMust(t, store)

		var allRevisions []*User
		err = session.Advanced().Revisions().GetFor(&allRevisions, "users/1")
		assert.NoError(t, err)
		assert.Equal(t, len(allRevisions), 4)

		names := collectUserNamesSorted(allRevisions)
		assert.Equal(t, names, []string{"user1", "user2", "user3", "user4"})

		var revisionsSkipFirst []*User
		err = session.Advanced().Revisions().GetForStartAt(&revisionsSkipFirst, "users/1", 1)
		assert.NoError(t, err)
		assert.Equal(t, len(revisionsSkipFirst), 3)
		names = collectUserNamesSorted(revisionsSkipFirst)
		assert.Equal(t, names, []string{"user1", "user2", "user3"})

		var revisionsSkipFirstTakeTwo []*User
		err = session.Advanced().Revisions().GetForPaged(&revisionsSkipFirstTakeTwo, "users/1", 1, 2)
		assert.NoError(t, err)
		assert.Equal(t, len(revisionsSkipFirstTakeTwo), 2)
		names = collectUserNamesSorted(revisionsSkipFirstTakeTwo)
		assert.Equal(t, names, []string{"user2", "user3"})

		allMetadata, err := session.Advanced().Revisions().GetMetadataFor("users/1")
		assert.NoError(t, err)
		assert.Equal(t, len(allMetadata), 4)

		metadataSkipFirst, err := session.Advanced().Revisions().GetMetadataForStartAt("users/1", 1)
		assert.NoError(t, err)
		assert.Equal(t, len(metadataSkipFirst), 3)

		metadataSkipFirstTakeTwo, err := session.Advanced().Revisions().GetMetadataForPaged("users/1", 1, 2)
		assert.NoError(t, err)
		assert.Equal(t, len(metadataSkipFirstTakeTwo), 2)

		dict := metadataSkipFirst[0]
		var changeVector string
		chvi, ok := dict.Get(ravendb.MetadataChangeVector)
		if ok {
			changeVector = chvi.(string)
		}
		var user *User
		err = session.Advanced().Revisions().Get(&user, changeVector)
		assert.NoError(t, err)
		assert.Equal(t, *user.Name, "user3")
		session.Close()
	}
}

func revisionsTestCanListRevisionsBin(t *testing.T, driver *RavenTestDriver) {

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	_, err = setupRevisions(store, false, 4)
	assert.NoError(t, err)

	createRevisions(t, store)

	{
		session := openSessionMust(t, store)

		user := &User{}
		user.setName("user1")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		err = session.DeleteByID("users/1", nil)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}
	revisionsBinEntryCommand := ravendb.NewGetRevisionsBinEntryCommand(math.MaxInt64, 20)
	err = store.GetRequestExecutor("").ExecuteCommand(revisionsBinEntryCommand, nil)
	assert.NoError(t, err)
	result := revisionsBinEntryCommand.Result
	assert.Equal(t, len(result.Results), 1)
	metaI := result.Results[0]["@metadata"]
	meta := metaI.(map[string]interface{})
	id, _ := jsonGetAsText(meta, "@id")
	assert.Equal(t, id, "users/1")
}

// for better code coverage
func goRevisionsTest(t *testing.T, driver *RavenTestDriver) {

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	_, err = setupRevisions(store, false, 4)
	assert.NoError(t, err)

	createRevisions(t, store)

	{
		session := openSessionMust(t, store)

		allMetadata, err := session.Advanced().Revisions().GetMetadataFor("users/1")
		assert.NoError(t, err)
		assert.Equal(t, len(allMetadata), 4)

		var changeVectors []string
		for _, dict := range allMetadata {
			var changeVector string
			chvi, ok := dict.Get(ravendb.MetadataChangeVector)
			if ok {
				changeVector = chvi.(string)
				changeVectors = append(changeVectors, changeVector)
			}
		}
		assert.Equal(t, len(changeVectors), 4)

		revisions := map[string]*User{}
		err = session.Advanced().Revisions().GetRevisions(revisions, changeVectors)
		assert.NoError(t, err)
		assert.Equal(t, len(revisions), 4)
		session.Close()
	}
}

func TestRevisions(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	revisionsTestRevisions(t, driver)
	// TODO: order might be different than Java
	revisionsTestCanListRevisionsBin(t, driver)

	goRevisionsTest(t, driver)
}

