package ravendb

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"runtime/debug"
	"sort"
	"strconv"
	"testing"

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

		err = session.StoreWithID(user, "users/1")
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
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		stream := bytes.NewBuffer([]byte{1, 2, 3})

		user := NewUser()
		user.setName("Fitzchak")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		err = session.advanced().attachments().storeEntity(user, "profile", stream, "image/png")
		assert.NoError(t, err)
		err = session.advanced().attachments().storeEntity(user, "other", stream, "")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.Error(t, err)
		_, ok := err.(*IllegalStateException)
		assert.True(t, ok)

		session.Close()
	}
}

func attachmentsSession_throwWhenTwoAttachmentsWithTheSameNameInSession(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		stream := bytes.NewBuffer([]byte{1, 2, 3})
		stream2 := bytes.NewBuffer([]byte{1, 2, 3, 4, 5})

		user := NewUser()
		user.setName("Fitzchak")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		err = session.advanced().attachments().storeEntity(user, "profile", stream, "image/png")
		assert.NoError(t, err)

		err = session.advanced().attachments().storeEntity(user, "profile", stream2, "")
		assert.Error(t, err)
		_, ok := err.(*IllegalStateException)
		assert.True(t, ok)

		session.Close()
	}
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
		err = session.StoreWithID(user, "users/1")
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
		err = session.StoreWithID(user, "users/1")

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
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user := NewUser()
		user.setName("Fitzchak")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		stream1 := bytes.NewBuffer([]byte{1, 2, 3})
		stream2 := bytes.NewBuffer([]byte{1, 2, 3, 4, 5, 6})

		err = session.advanced().attachments().storeEntity(user, "file1", stream1, "image/png")
		assert.NoError(t, err)
		err = session.advanced().attachments().storeEntity(user, "file2", stream2, "image/png")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	op := NewDeleteAttachmentOperation("users/1", "file2", nil)
	err = store.operations().send(op)
	assert.NoError(t, err)

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
		assert.Equal(t, len(attachments), 1)

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

func attachmentsSession_getAttachmentReleasesResources(t *testing.T) {
	count := 30
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user := NewUser()
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	for i := 0; i < count; i++ {
		session := openSessionMust(t, store)

		stream1 := bytes.NewBuffer([]byte{1, 2, 3})
		err = session.advanced().attachments().store("users/1", "file"+strconv.Itoa(i), stream1, "image/png")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	for i := 0; i < count; i++ {
		session := openSessionMust(t, store)
		result, err := session.advanced().attachments().get("users/1", "file"+strconv.Itoa(i))
		assert.NoError(t, err)
		// don't read data as it marks entity as consumed
		result.Close()
		session.Close()
	}
}

func attachmentsSession_deleteDocumentAndThanItsAttachments_ThisIsNoOpButShouldBeSupported(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user := NewUser()
		user.setName("Fitzchak")
		err = session.StoreWithID(user, "users/1")
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

		userI, err := session.load(getTypeOf(&User{}), "users/1")
		assert.NoError(t, err)

		err = session.DeleteEntity(userI)
		assert.NoError(t, err)
		err = session.advanced().attachments().deleteEntity(userI, "file")
		assert.NoError(t, err)
		err = session.advanced().attachments().deleteEntity(userI, "file") // this should be no-op
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}
}

func attachmentsSession_deleteDocumentByCommandAndThanItsAttachments_ThisIsNoOpButShouldBeSupported(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("Fitzchak")
		err = session.StoreWithID(user, "users/1")
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
		err = session.StoreWithID(user, "users/1")
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
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		stream := bytes.NewBuffer([]byte{1, 2, 3})

		user := NewUser()
		user.setName("Fitzchak")

		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		err = session.advanced().attachments().store("users/1", "profile", stream, "image/png")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		ok, err := session.advanced().attachments().exists("users/1", "profile")
		assert.NoError(t, err)
		assert.True(t, ok)

		ok, err = session.advanced().attachments().exists("users/1", "background-photo")
		assert.NoError(t, err)
		assert.False(t, ok)

		ok, err = session.advanced().attachments().exists("users/2", "profile")
		assert.NoError(t, err)
		assert.False(t, ok)

		session.Close()
	}
}

func TestAttachmentsSession(t *testing.T) {
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

	// TODO: those tests are flaky. Not often but they sometimes fail

	// matches order of Java tests

	// TODO: re-eneable when not flaky
	if gEnableFlakyTests {
		attachmentsSession_putAttachments(t)
	}
	attachmentsSession_putDocumentAndAttachmentAndDeleteShouldThrow(t)

	// TODO: re-eneable when not flaky
	if gEnableFlakyTests {
		attachmentsSession_getAttachmentNames(t)
	}
	attachmentsSession_deleteDocumentByCommandAndThanItsAttachments_ThisIsNoOpButShouldBeSupported(t)
	attachmentsSession_deleteAttachments(t)
	attachmentsSession_attachmentExists(t)
	attachmentsSession_throwWhenTwoAttachmentsWithTheSameNameInSession(t)
	attachmentsSession_deleteDocumentAndThanItsAttachments_ThisIsNoOpButShouldBeSupported(t)
	attachmentsSession_throwIfStreamIsUseTwice(t)
	attachmentsSession_getAttachmentReleasesResources(t)
	attachmentsSession_deleteAttachmentsUsingCommand(t)
}
