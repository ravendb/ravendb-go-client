package ravendb

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"runtime/debug"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func attachmentsRevisions_putAttachments(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		_, err = setupRevisions(store, false, 4)
		assert.NoError(t, err)

		names := createDocumentWithAttachments(t, store)
		{
			f := func(t *testing.T, session *DocumentSession, revisions []*User) {
				assertRevisionAttachments(t, names, 3, revisions[0], session)
				assertRevisionAttachments(t, names, 2, revisions[1], session)
				assertRevisionAttachments(t, names, 1, revisions[2], session)
				assertNoRevisionAttachment(t, revisions[3], session, false)
			}
			assertRevisions(t, store, names, f, 9)
		}
		{
			// Delete document should delete all the attachments
			f := func(t *testing.T, session *DocumentSession, revisions []*User) {
				assertNoRevisionAttachment(t, revisions[0], session, true)
				assertRevisionAttachments(t, names, 3, revisions[1], session)
				assertRevisionAttachments(t, names, 2, revisions[2], session)
				assertRevisionAttachments(t, names, 1, revisions[3], session)
			}

			cmd := NewDeleteDocumentCommand("users/1", nil)
			err = store.getRequestExecutor().executeCommand(cmd)
			assert.NoError(t, err)
			assertRevisions2(t, store, names, f, 6, 0, 3)
		}

		{
			// Create another revision which should delete old revision
			session := openSessionMust(t, store)
			// This will delete the revision #1 which is without attachment
			user := NewUser()
			user.setName("Fitzchak 2")
			err = session.StoreWithID(user, "users/1")
			assert.NoError(t, err)
			err = session.SaveChanges()
			assert.NoError(t, err)
			session.Close()
		}

		{
			f := func(t *testing.T, session *DocumentSession, revisions []*User) {
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
			user := NewUser()
			user.setName("Fitzchak 3")
			err = session.StoreWithID(user, "users/1")
			assert.NoError(t, err)
			err = session.SaveChanges()
			assert.NoError(t, err)
			session.Close()
		}

		{
			f := func(t *testing.T, session *DocumentSession, revisions []*User) {
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
			user := NewUser()
			user.setName("Fitzchak 4")
			err = session.StoreWithID(user, "users/1")
			assert.NoError(t, err)
			err = session.SaveChanges()
			assert.NoError(t, err)
			session.Close()
		}

		{
			f := func(t *testing.T, session *DocumentSession, revisions []*User) {
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
			user := NewUser()
			user.setName("Fitzchak 5")
			err = session.StoreWithID(user, "users/1")
			assert.NoError(t, err)
			err = session.SaveChanges()
			assert.NoError(t, err)
			session.Close()
		}

		{
			f := func(t *testing.T, session *DocumentSession, revisions []*User) {
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

func attachmentsRevisions_attachmentRevision(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		_, err = setupRevisions(store, false, 4)
		assert.NoError(t, err)

		createDocumentWithAttachments(t, store)

		{
			session := openSessionMust(t, store)
			bais := bytes.NewBuffer([]byte{5, 4, 3, 2, 1})
			err = session.advanced().attachments().store("users/1", "profile.png", bais, "")
			assert.NoError(t, err)

			err = session.SaveChanges()
			assert.NoError(t, err)
			session.Close()
		}

		{
			session := openSessionMust(t, store)
			revisionsI, err := session.advanced().revisions().getFor(getTypeOf(&User{}), "users/1")
			assert.NoError(t, err)

			// TODO: could be done with reflection
			n := len(revisionsI)
			revisions := make([]*User, n, n)
			for i, revI := range revisionsI {
				revisions[i] = revI.(*User)
			}

			rev := revisions[1]
			changeVector, err := session.advanced().getChangeVectorFor(rev)
			assert.NoError(t, err)

			{
				revision, err := session.advanced().attachments().getRevision("users/1", "profile.png", changeVector)
				assert.NoError(t, err)
				r := revision.getData()
				bytes, err := ioutil.ReadAll(r)
				assert.NoError(t, err)
				assert.Equal(t, len(bytes), 3)
				assert.Equal(t, bytes, []byte{1, 2, 3})
			}
			session.Close()
		}
	}
}

func createDocumentWithAttachments(t *testing.T, store *DocumentStore) []string {
	var err error
	{
		session := openSessionMust(t, store)

		user := NewUser()
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
		op := NewPutAttachmentOperation("users/1", names[0], profileStream, "image/png", nil)
		// TODO this test is flaky. Sometimes it works, sometimes it doesn't
		// even though the data sent on wire seem to be the same
		err = store.operations().send(op)

		assert.NoError(t, err)

		result := op.Command.Result
		s := *result.getChangeVector()
		assert.True(t, strings.Contains(s, "A:3"))
		assert.Equal(t, result.getName(), names[0])
		assert.Equal(t, result.getDocumentId(), "users/1")
		assert.Equal(t, result.getContentType(), "image/png")
	}

	{
		backgroundStream := bytes.NewReader([]byte{10, 20, 30, 40, 50})
		op := NewPutAttachmentOperation("users/1", names[1], backgroundStream, "ImGgE/jPeG", nil)
		err = store.operations().send(op)
		assert.NoError(t, err)
		result := op.Command.Result
		s := *result.getChangeVector()
		assert.True(t, strings.Contains(s, "A:7"))
		assert.Equal(t, result.getName(), names[1])
		assert.Equal(t, result.getDocumentId(), "users/1")
		assert.Equal(t, result.getContentType(), "ImGgE/jPeG")
	}
	{
		fileStream := bytes.NewReader([]byte{1, 2, 3, 4, 5})
		op := NewPutAttachmentOperation("users/1", names[2], fileStream, "", nil)
		err = store.operations().send(op)
		assert.NoError(t, err)
		result := op.Command.Result
		s := *result.getChangeVector()
		assert.True(t, strings.Contains(s, "A:12"))
		assert.Equal(t, result.getName(), names[2])
		assert.Equal(t, result.getDocumentId(), "users/1")
		assert.Equal(t, result.getContentType(), "")
	}
	return names
}

func assertRevisions(t *testing.T, store *DocumentStore, names []string, assertAction func(*testing.T, *DocumentSession, []*User), expectedCountOfAttachments int) {
	assertRevisions2(t, store, names, assertAction, expectedCountOfAttachments, 1, 3)
}

func assertRevisions2(t *testing.T, store *DocumentStore, names []string, assertAction func(*testing.T, *DocumentSession, []*User), expectedCountOfAttachments int, expectedCountOfDocuments int, expectedCountOfUniqueAttachments int) {
	op := NewGetStatisticsOperation()
	err := store.maintenance().send(op)
	assert.NoError(t, err)
	statistics := op.Command.Result

	assert.Equal(t, statistics.getCountOfAttachments(), expectedCountOfAttachments)

	assert.Equal(t, statistics.getCountOfUniqueAttachments(), expectedCountOfUniqueAttachments)

	assert.Equal(t, statistics.getCountOfRevisionDocuments(), 4)

	assert.Equal(t, statistics.getCountOfDocuments(), expectedCountOfDocuments)

	assert.Equal(t, statistics.getCountOfIndexes(), 0)

	{
		session := openSessionMust(t, store)
		revisionsI, err := session.advanced().revisions().getFor(getTypeOf(&User{}), "users/1")
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

func assertNoRevisionAttachment(t *testing.T, revision *User, session *DocumentSession, isDeleteRevision bool) {
	metadata, err := session.advanced().getMetadataFor(revision)
	assert.NoError(t, err)

	if isDeleteRevision {
		v, _ := metadata.get(Constants_Documents_Metadata_FLAGS)
		s := v.(string)
		assert.True(t, strings.Contains(s, "HasRevisions"))
		assert.True(t, strings.Contains(s, "DeleteRevision"))
	} else {
		v, _ := metadata.get(Constants_Documents_Metadata_FLAGS)
		s := v.(string)
		assert.True(t, strings.Contains(s, "HasRevisions"))
		assert.True(t, strings.Contains(s, "Revision"))
	}

	hasIt := metadata.containsKey(Constants_Documents_Metadata_ATTACHMENTS)
	assert.False(t, hasIt)
}

func assertRevisionAttachments(t *testing.T, names []string, expectedCount int, revision *User, session *DocumentSession) {
	metadata, err := session.advanced().getMetadataFor(revision)
	assert.NoError(t, err)
	v, _ := metadata.get(Constants_Documents_Metadata_FLAGS)
	s := v.(string)
	assert.True(t, strings.Contains(s, "HasRevisions"))
	assert.True(t, strings.Contains(s, "Revision"))
	assert.True(t, strings.Contains(s, "Revision"))

	attachments := metadata.getObjects(Constants_Documents_Metadata_ATTACHMENTS)
	assert.Equal(t, len(attachments), expectedCount)

	// Note: unlike Java, compare them after sorting
	attachmentNames := make([]string, expectedCount, expectedCount)
	for i := 0; i < expectedCount; i++ {
		attachment := attachments[i]
		aname, ok := attachment.get("Name")
		anameStr, ok := aname.(string)
		assert.True(t, ok)
		attachmentNames[i] = anameStr
	}

	orderedNames := stringArrayCopy(names)
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

	destroyDriver := createTestDriver(t)
	defer func() {
		r := recover()
		destroyDriver()
		if r != nil {
			fmt.Printf("Panic: '%v'\n", r)
			debug.PrintStack()
			t.Fail()
		}
	}()

	// matches order of Java tests

	// TODO: this test is flaky. See bugs.txt
	// Note: it also fails in Java on mac pro
	//attachmentsRevisions_putAttachments(t)
	//attachmentsRevisions_attachmentRevision(t)

}
