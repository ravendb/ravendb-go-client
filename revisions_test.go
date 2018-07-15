package ravendb

import (
	"sort"
	"strconv"
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func collectUserNamesSorted(a []interface{}) []string {
	var names []string
	for _, v := range a {
		user := v.(*User)
		names = append(names, *user.getName())
	}
	sort.Strings(names)
	return names
}

func revisionsTest_revisions(t *testing.T) {

	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	_, err = setupRevisions(store, false, 4)
	assert.NoError(t, err)

	for i := 0; i < 4; i++ {
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("user" + strconv.Itoa(i+1))
		err = session.StoreEntityWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)

		allRevisions, err := session.advanced().revisions().getFor(getTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		assert.Equal(t, len(allRevisions), 4)

		names := collectUserNamesSorted(allRevisions)
		assert.Equal(t, names, []string{"user1", "user2", "user3", "user4"})

		revisionsSkipFirst, err := session.advanced().revisions().getForStartAt(getTypeOf(&User{}), "users/1", 1)
		assert.NoError(t, err)
		assert.Equal(t, len(revisionsSkipFirst), 3)
		names = collectUserNamesSorted(revisionsSkipFirst)
		assert.Equal(t, names, []string{"user1", "user2", "user3"})

		revisionsSkipFirstTakeTwo, err := session.advanced().revisions().getForPaged(getTypeOf(&User{}), "users/1", 1, 2)
		assert.NoError(t, err)
		assert.Equal(t, len(revisionsSkipFirstTakeTwo), 2)
		names = collectUserNamesSorted(revisionsSkipFirstTakeTwo)
		assert.Equal(t, names, []string{"user2", "user3"})

		allMetadata, err := session.advanced().revisions().getMetadataFor("users/1")
		assert.NoError(t, err)
		assert.Equal(t, len(allMetadata), 4)

		metadataSkipFirst, err := session.advanced().revisions().getMetadataForStartAt("users/1", 1)
		assert.NoError(t, err)
		assert.Equal(t, len(metadataSkipFirst), 3)

		metadataSkipFirstTakeTwo, err := session.advanced().revisions().getMetadataForPaged("users/1", 1, 2)
		assert.NoError(t, err)
		assert.Equal(t, len(metadataSkipFirstTakeTwo), 2)

		dict := metadataSkipFirst[0]
		var changeVector string
		chvi, ok := dict.get(Constants_Documents_Metadata_CHANGE_VECTOR)
		if ok {
			changeVector = chvi.(string)
		}
		userI, err := session.advanced().revisions().get(getTypeOf(&User{}), changeVector)
		assert.NoError(t, err)
		user := userI.(*User)
		assert.Equal(t, *user.getName(), "user3")
		session.Close()
	}
}

func TestRevisions(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_revisions_go.txt")
	}

	createTestDriver()
	defer deleteTestDriver()

	revisionsTest_revisions(t)
}
