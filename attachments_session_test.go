package ravendb

import (
	"bytes"
	"io/ioutil"
	"sort"
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func attachmentsSession_putAttachments(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()
	names := []string{"profile.png", "background-photo.jpg", "fileNAME_#$1^%_בעברית.txt"}

	{
		session := openSessionMust(t, store)
		profileStream := bytes.NewBuffer([]byte{1, 2, 3})
		backgroundStream := bytes.NewBuffer([]byte{10, 20, 30, 40, 50})
		fileStream := bytes.NewBuffer([]byte{1, 2, 3, 4, 5})

		user := NewUser()
		user.setName("Fitzchak")

		err = session.StoreEntityWithID(user, "users/1")
		assert.NoError(t, err)

		err = session.advanced().attachments().store("users/1", names[0], profileStream, "image/png")
		assert.NoError(t, err)
		err = session.advanced().attachments().storeEntity(user, names[1], backgroundStream, "ImGgE/jPeG")
		assert.NoError(t, err)
		err = session.advanced().attachments().storeEntity(user, names[2], fileStream, "")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		userI, err := session.load(getTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		user := userI.(*User)
		metadata, err := session.advanced().getMetadataFor(user)
		assert.NoError(t, err)
		v, ok := metadata.get(Constants_Documents_Metadata_FLAGS)
		assert.True(t, ok)
		vStr := v.(string)
		assert.Equal(t, vStr, "HasAttachments")

		attachmentsI, ok := metadata.get(Constants_Documents_Metadata_ATTACHMENTS)
		assert.True(t, ok)
		attachments := attachmentsI.([]Object)
		assert.Equal(t, len(attachments), 3)

		sort.Strings(names)
		var gotNames []string
		for _, v := range attachments {
			attachment := v.(*IMetadataDictionary)
			name, ok := attachment.get("Name")
			assert.True(t, ok)
			gotNames = append(gotNames, name.(string))
		}
		sort.Strings(gotNames)
		assert.Equal(t, names, gotNames)
		session.Close()
	}
}

func attachmentsSession_throwIfStreamIsUseTwice(t *testing.T) {
}

func attachmentsSession_throwWhenTwoAttachmentsWithTheSameNameInSession(t *testing.T) {
}

func attachmentsSession_putDocumentAndAttachmentAndDeleteShouldThrow(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		profileStream := bytes.NewBuffer([]byte{1, 2, 3})
		user := NewUser()
		user.setName("Fitzchak")
		err = session.StoreEntityWithID(user, "users/1")
		assert.NoError(t, err)

		err = session.advanced().attachments().storeEntity(user, "profile.png", profileStream, "image/png")
		assert.NoError(t, err)

		err = session.DeleteEntity(user)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.Error(t, err)
		_, ok := err.(*IllegalStateException)
		assert.True(t, ok)

		session.Close()
	}
}

func attachmentsSession_deleteAttachments(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user := NewUser()
		user.setName("Fitzchak")
		err = session.StoreEntityWithID(user, "users/1")

		stream1 := bytes.NewBuffer([]byte{1, 2, 3})
		stream2 := bytes.NewBuffer([]byte{1, 2, 3, 4, 5, 6})
		stream3 := bytes.NewBuffer([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})
		stream4 := bytes.NewBuffer([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})

		err = session.advanced().attachments().storeEntity(user, "file1", stream1, "image/png")
		assert.NoError(t, err)
		err = session.advanced().attachments().storeEntity(user, "file2", stream2, "image/png")
		assert.NoError(t, err)
		err = session.advanced().attachments().storeEntity(user, "file3", stream3, "image/png")
		assert.NoError(t, err)
		err = session.advanced().attachments().storeEntity(user, "file4", stream4, "image/png")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		userI, err := session.load(getTypeOf(&User{}), "users/1")
		assert.NoError(t, err)

		// test get attachment by its name
		{
			attachmentResult, err := session.advanced().attachments().get("users/1", "file2")
			assert.NoError(t, err)
			name := attachmentResult.getDetails().getName()
			assert.Equal(t, name, "file2")
			attachmentResult.Close()
		}

		err = session.advanced().attachments().delete("users/1", "file2")
		assert.NoError(t, err)
		err = session.advanced().attachments().deleteEntity(userI, "file4")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		userI, err := session.load(getTypeOf(&User{}), "users/1")
		assert.NoError(t, err)

		metadata, err := session.advanced().getMetadataFor(userI)
		assert.NoError(t, err)

		v, ok := metadata.get(Constants_Documents_Metadata_FLAGS)
		assert.True(t, ok)
		assert.Equal(t, v, "HasAttachments")

		attachmentsI, ok := metadata.get(Constants_Documents_Metadata_ATTACHMENTS)
		assert.True(t, ok)
		attachments := attachmentsI.([]Object)

		assert.Equal(t, len(attachments), 2)

		{
			result, err := session.advanced().attachments().get("users/1", "file1")
			assert.NoError(t, err)
			r := result.getData()
			file1Bytes, err := ioutil.ReadAll(r)
			assert.NoError(t, err)

			assert.Equal(t, len(file1Bytes), 3)

			result.Close()
		}
		session.Close()
	}
}

func attachmentsSession_deleteAttachmentsUsingCommand(t *testing.T) {
}

func attachmentsSession_getAttachmentReleasesResources(t *testing.T) {
}

func attachmentsSession_deleteDocumentAndThanItsAttachments_ThisIsNoOpButShouldBeSupported(t *testing.T) {
}

func attachmentsSession_deleteDocumentByCommandAndThanItsAttachments_ThisIsNoOpButShouldBeSupported(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("Fitzchak")
		err = session.StoreEntityWithID(user, "users/1")
		assert.NoError(t, err)

		stream := bytes.NewBuffer([]byte{1, 2, 3})

		err = session.advanced().attachments().storeEntity(user, "file", stream, "image/png")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		cd := NewDeleteCommandData("users/1", nil)
		session.advanced().Defer(cd)
		err = session.advanced().attachments().delete("users/1", "file")
		assert.NoError(t, err)
		err = session.advanced().attachments().delete("users/1", "file")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}
}

func attachmentsSession_getAttachmentNames(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	names := []string{"profile.png"}

	{
		session := openSessionMust(t, store)

		profileStream := bytes.NewBuffer([]byte{1, 2, 3})

		user := NewUser()
		user.setName("Fitzchak")
		err = session.StoreEntityWithID(user, "users/1")
		assert.NoError(t, err)

		err = session.advanced().attachments().store("users/1", names[0], profileStream, "image/png")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		userI, err := session.load(getTypeOf(&User{}), "users/1")
		assert.NoError(t, err)

		attachments, err := session.advanced().attachments().getNames(userI)
		assert.NoError(t, err)

		assert.Equal(t, len(attachments), 1)

		attachment := attachments[0]
		assert.Equal(t, attachment.getContentType(), "image/png")
		assert.Equal(t, attachment.getName(), names[0])
		assert.Equal(t, attachment.getSize(), int64(3))

		session.Close()
	}
}

func attachmentsSession_attachmentExists(t *testing.T) {
}

func TestAttachmentsSession(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_attachments_session_go.txt")
	}

	if true {
		dumpFailedHTTP = true
	}

	createTestDriver()
	defer deleteTestDriver()

	// matches order of Java tests
	attachmentsSession_putAttachments(t)
	attachmentsSession_putDocumentAndAttachmentAndDeleteShouldThrow(t)
	attachmentsSession_getAttachmentNames(t)
	attachmentsSession_deleteDocumentByCommandAndThanItsAttachments_ThisIsNoOpButShouldBeSupported(t)
	attachmentsSession_deleteAttachments(t)

	attachmentsSession_attachmentExists(t)
	attachmentsSession_throwWhenTwoAttachmentsWithTheSameNameInSession(t)
	attachmentsSession_deleteDocumentAndThanItsAttachments_ThisIsNoOpButShouldBeSupported(t)
	attachmentsSession_throwIfStreamIsUseTwice(t)
	attachmentsSession_getAttachmentReleasesResources(t)
	attachmentsSession_deleteAttachmentsUsingCommand(t)
}
