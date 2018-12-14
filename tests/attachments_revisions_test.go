package tests

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func attachmentsRevisions_putAttachments(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		_, err = setupRevisions(store, false, 4)
		assert.NoError(t, err)

		names := createDocumentWithAttachments(t, store)
		{
			f := func(t *testing.T, session *ravendb.DocumentSession, revisions []*User) {
				assertRevisionAttachments(t, names, 3, revisions[0], session)
				assertRevisionAttachments(t, names, 2, revisions[1], session)
				assertRevisionAttachments(t, names, 1, revisions[2], session)
				assertNoRevisionAttachment(t, revisions[3], session, false)
			}
			assertRevisions(t, store, names, f, 9)
		}
		{
			// Delete document should delete all the attachments
			f := func(t *testing.T, session *ravendb.DocumentSession, revisions []*User) {
				assertNoRevisionAttachment(t, revisions[0], session, true)
				assertRevisionAttachments(t, names, 3, revisions[1], session)
				assertRevisionAttachments(t, names, 2, revisions[2], session)
				assertRevisionAttachments(t, names, 1, revisions[3], session)
			}

			cmd := ravendb.NewDeleteDocumentCommand("users/1", nil)
			err = store.GetRequestExecutor().ExecuteCommand(cmd)
			assert.NoError(t, err)
			assertRevisions2(t, store, names, f, 6, 0, 3)
		}

		{
			// Create another revision which should delete old revision
			session := openSessionMust(t, store)
			// This will delete the revision #1 which is without attachment
			user := User{}
			user.setName("Fitzchak 2")
			err = session.StoreWithID(user, "users/1")
			assert.NoError(t, err)
			err = session.SaveChanges()
			assert.NoError(t, err)
			session.Close()
		}

		{
			f := func(t *testing.T, session *ravendb.DocumentSession, revisions []*User) {
				// This will delete the revision #2 which is with attachment
				assertNoRevisionAttachment(t, revisions[0], session, false)
				assertNoRevisionAttachment(t, revisions[1], session, true)
				assertRevisionAttachments(t, names, 3, revisions[2], session)
				assertRevisionAttachments(t, names, 2, revisions[3], session)
			}
			assertRevisions2(t, store, names, f, 5, 1, 3)
		}

		{
			session := openSessionMust(t, store)
			// This will delete the revision #2 which is with attachment
			user := &User{}
			user.setName("Fitzchak 3")
			err = session.StoreWithID(user, "users/1")
			assert.NoError(t, err)
			err = session.SaveChanges()
			assert.NoError(t, err)
			session.Close()
		}

		{
			f := func(t *testing.T, session *ravendb.DocumentSession, revisions []*User) {
				// This will delete the revision #2 which is with attachment
				assertNoRevisionAttachment(t, revisions[0], session, false)
				assertNoRevisionAttachment(t, revisions[1], session, false)
				assertNoRevisionAttachment(t, revisions[2], session, true)
				assertRevisionAttachments(t, names, 3, revisions[3], session)
			}
			assertRevisions2(t, store, names, f, 3, 1, 3)
		}

		{
			session := openSessionMust(t, store)
			// This will delete the revision #3 which is with attachment
			user := &User{}
			user.setName("Fitzchak 4")
			err = session.StoreWithID(user, "users/1")
			assert.NoError(t, err)
			err = session.SaveChanges()
			assert.NoError(t, err)
			session.Close()
		}

		{
			f := func(t *testing.T, session *ravendb.DocumentSession, revisions []*User) {
				// This will delete the revision #2 which is with attachment
				assertNoRevisionAttachment(t, revisions[0], session, false)
				assertNoRevisionAttachment(t, revisions[1], session, false)
				assertNoRevisionAttachment(t, revisions[2], session, false)
				assertNoRevisionAttachment(t, revisions[3], session, true)
			}
			assertRevisions2(t, store, names, f, 0, 1, 0)
		}

		{
			session := openSessionMust(t, store)
			// This will delete the revision #4 which is with attachment
			user := &User{}
			user.setName("Fitzchak 5")
			err = session.StoreWithID(user, "users/1")
			assert.NoError(t, err)
			err = session.SaveChanges()
			assert.NoError(t, err)
			session.Close()
		}

		{
			f := func(t *testing.T, session *ravendb.DocumentSession, revisions []*User) {
				// This will delete the revision #2 which is with attachment
				assertNoRevisionAttachment(t, revisions[0], session, false)
				assertNoRevisionAttachment(t, revisions[1], session, false)
				assertNoRevisionAttachment(t, revisions[2], session, false)
				assertNoRevisionAttachment(t, revisions[3], session, false)
			}
			assertRevisions2(t, store, names, f, 0, 1, 0)
		}
	}
}

func attachmentsRevisions_attachmentRevision(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		_, err = setupRevisions(store, false, 4)
		assert.NoError(t, err)

		createDocumentWithAttachments(t, store)

		{
			session := openSessionMust(t, store)
			bais := bytes.NewBuffer([]byte{5, 4, 3, 2, 1})
			err = session.Advanced().Attachments().Store("users/1", "profile.png", bais, "")
			assert.NoError(t, err)

			err = session.SaveChanges()
			assert.NoError(t, err)
			session.Close()
		}

		{
			session := openSessionMust(t, store)
			revisionsI, err := session.Advanced().Revisions().GetFor(reflect.TypeOf(&User{}), "users/1")
			assert.NoError(t, err)

			// TODO: could be done with reflection
			n := len(revisionsI)
			revisions := make([]*User, n, n)
			for i, revI := range revisionsI {
				revisions[i] = revI.(*User)
			}

			rev := revisions[1]
			changeVector, err := session.Advanced().GetChangeVectorFor(rev)
			assert.NoError(t, err)

			{
				revision, err := session.Advanced().Attachments().GetRevision("users/1", "profile.png", changeVector)
				assert.NoError(t, err)
				r := revision.GetData()
				bytes, err := ioutil.ReadAll(r)
				assert.NoError(t, err)
				assert.Equal(t, len(bytes), 3)
				assert.Equal(t, bytes, []byte{1, 2, 3})
			}
			session.Close()
		}
	}
}

func createDocumentWithAttachments(t *testing.T, store *ravendb.DocumentStore) []string {
	var err error
	{
		session := openSessionMust(t, store)

		user := &User{}
		user.setName("Fitzchak")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	names := []string{
		"profile.png",
		"background-photo.jpg",
		"fileNAME_#$1^%_בעברית.txt",
	}

	{
		profileStream := bytes.NewBuffer([]byte{1, 2, 3})
		op := ravendb.NewPutAttachmentOperation("users/1", names[0], profileStream, "image/png", nil)
		// TODO this test is flaky. Sometimes it works, sometimes it doesn't
		// even though the data sent on wire seem to be the same
		err = store.Operations().Send(op)

		assert.NoError(t, err)

		result := op.Command.Result
		s := *result.ChangeVector
		assert.True(t, strings.Contains(s, "A:3"))
		assert.Equal(t, result.Name, names[0])
		assert.Equal(t, result.DocumentID, "users/1")
		assert.Equal(t, result.ContentType, "image/png")
	}

	{
		backgroundStream := bytes.NewReader([]byte{10, 20, 30, 40, 50})
		op := ravendb.NewPutAttachmentOperation("users/1", names[1], backgroundStream, "ImGgE/jPeG", nil)
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		result := op.Command.Result
		s := *result.ChangeVector
		assert.True(t, strings.Contains(s, "A:7"))
		assert.Equal(t, result.Name, names[1])
		assert.Equal(t, result.DocumentID, "users/1")
		assert.Equal(t, result.ContentType, "ImGgE/jPeG")
	}
	{
		fileStream := bytes.NewReader([]byte{1, 2, 3, 4, 5})
		op := ravendb.NewPutAttachmentOperation("users/1", names[2], fileStream, "", nil)
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		result := op.Command.Result
		s := *result.ChangeVector
		assert.True(t, strings.Contains(s, "A:12"))
		assert.Equal(t, result.Name, names[2])
		assert.Equal(t, result.DocumentID, "users/1")
		assert.Equal(t, result.ContentType, "")
	}
	return names
}

func assertRevisions(t *testing.T, store *ravendb.DocumentStore, names []string, assertAction func(*testing.T, *ravendb.DocumentSession, []*User), expectedCountOfAttachments int) {
	assertRevisions2(t, store, names, assertAction, expectedCountOfAttachments, 1, 3)
}

func assertRevisions2(t *testing.T, store *ravendb.DocumentStore, names []string, assertAction func(*testing.T, *ravendb.DocumentSession, []*User), expectedCountOfAttachments int, expectedCountOfDocuments int, expectedCountOfUniqueAttachments int) {
	op := ravendb.NewGetStatisticsOperation()
	err := store.Maintenance().Send(op)
	assert.NoError(t, err)
	statistics := op.Command.Result

	assert.Equal(t, statistics.CountOfAttachments, expectedCountOfAttachments)

	assert.Equal(t, statistics.CountOfUniqueAttachments, expectedCountOfUniqueAttachments)

	assert.Equal(t, statistics.CountOfRevisionDocuments, 4)

	assert.Equal(t, statistics.CountOfDocuments, expectedCountOfDocuments)

	assert.Equal(t, statistics.CountOfIndexes, 0)

	{
		session := openSessionMust(t, store)
		revisionsI, err := session.Advanced().Revisions().GetFor(reflect.TypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		n := len(revisionsI)
		assert.Equal(t, n, 4)
		revisions := make([]*User, n, n)
		for i, v := range revisionsI {
			revisions[i] = v.(*User)
		}
		assertAction(t, session, revisions)

		session.Close()
	}
}

func assertNoRevisionAttachment(t *testing.T, revision *User, session *ravendb.DocumentSession, isDeleteRevision bool) {
	metadata, err := session.Advanced().GetMetadataFor(revision)
	assert.NoError(t, err)

	if isDeleteRevision {
		v, _ := metadata.Get(ravendb.Constants_Documents_Metadata_FLAGS)
		s := v.(string)
		assert.True(t, strings.Contains(s, "HasRevisions"))
		assert.True(t, strings.Contains(s, "DeleteRevision"))
	} else {
		v, _ := metadata.Get(ravendb.Constants_Documents_Metadata_FLAGS)
		s := v.(string)
		assert.True(t, strings.Contains(s, "HasRevisions"))
		assert.True(t, strings.Contains(s, "Revision"))
	}

	hasIt := metadata.ContainsKey(ravendb.Constants_Documents_Metadata_ATTACHMENTS)
	assert.False(t, hasIt)
}

func assertRevisionAttachments(t *testing.T, names []string, expectedCount int, revision *User, session *ravendb.DocumentSession) {
	metadata, err := session.Advanced().GetMetadataFor(revision)
	assert.NoError(t, err)
	v, _ := metadata.Get(ravendb.Constants_Documents_Metadata_FLAGS)
	s := v.(string)
	assert.True(t, strings.Contains(s, "HasRevisions"))
	assert.True(t, strings.Contains(s, "Revision"))
	assert.True(t, strings.Contains(s, "Revision"))

	attachments := metadata.GetObjects(ravendb.Constants_Documents_Metadata_ATTACHMENTS)
	assert.Equal(t, len(attachments), expectedCount)

	// Note: unlike Java, compare them after sorting
	attachmentNames := make([]string, expectedCount, expectedCount)
	for i := 0; i < expectedCount; i++ {
		attachment := attachments[i]
		aname, ok := attachment.Get("Name")
		assert.True(t, ok)
		anameStr, ok := aname.(string)
		assert.True(t, ok)
		attachmentNames[i] = anameStr
	}

	orderedNames := ravendb.StringArrayCopy(names)
	if len(orderedNames) > expectedCount {
		orderedNames = orderedNames[:expectedCount]
	}
	sort.Strings(orderedNames)
	sort.Strings(attachmentNames)
	assert.Equal(t, orderedNames, attachmentNames)
}

func TestAttachmentsRevisions(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests

	// TODO: this test is flaky. See bugs.txt
	// Note: it also fails in Java on mac pro
	// The bytes sent seem to be exactly the same, Go fails with EOF
	// Is it issue with not closing the request?
	if ravendb.EnableFlakyTests {
		attachmentsRevisions_putAttachments(t, driver)
	}
	if ravendb.EnableFlakyTests {
		attachmentsRevisions_attachmentRevision(t, driver)
	}
}
