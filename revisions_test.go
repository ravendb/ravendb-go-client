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
	}

	{
		session := openSessionMust(t, store)
		session.advanced()

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

	}
}

/*
	List<User> revisionsSkipFirstTakeTwo = session.advanced().revisions().getFor(User.class, "users/1", 1, 2);
	assertThat(revisionsSkipFirstTakeTwo)
			.hasSize(2);
	assertThat(revisionsSkipFirstTakeTwo.stream().map(x -> x.getName()).collect(Collectors.toList()))
			.containsSequence("user3", "user2" );

	List<MetadataAsDictionary> allMetadata = session.advanced().revisions().getMetadataFor("users/1");
	assertThat(allMetadata)
			.hasSize(4);

	List<MetadataAsDictionary> metadataSkipFirst = session.advanced().revisions().getMetadataFor("users/1", 1);
	assertThat(metadataSkipFirst)
			.hasSize(3);

	List<MetadataAsDictionary> metadataSkipFirstTakeTwo = session.advanced().revisions().getMetadataFor("users/1", 1, 2);
	assertThat(metadataSkipFirstTakeTwo)
			.hasSize(2);


	User user = session.advanced().revisions().get(User.class, (String) metadataSkipFirst.get(0).get(Constants.Documents.Metadata.CHANGE_VECTOR));
	assertThat(user.getName())
			.isEqualTo("user3");
*/

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
