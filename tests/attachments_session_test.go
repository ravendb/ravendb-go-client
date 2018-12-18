package tests

import (
	"bytes"
	"io/ioutil"
	"sort"
	"strconv"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func attachmentsSession_putAttachments(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()
	names := []string{"profile.png", "background-photo.jpg", "fileNAME_#$1^%_בעברית.txt"}

	{
		session := openSessionMust(t, store)
		profileStream := bytes.NewBuffer([]byte{1, 2, 3})
		backgroundStream := bytes.NewBuffer([]byte{10, 20, 30, 40, 50})
		fileStream := bytes.NewBuffer([]byte{1, 2, 3, 4, 5})

		user := &User{}
		user.setName("Fitzchak")

		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		err = session.Advanced().Attachments().Store("users/1", names[0], profileStream, "image/png")
		assert.NoError(t, err)
		err = session.Advanced().Attachments().StoreEntity(user, names[1], backgroundStream, "ImGgE/jPeG")
		assert.NoError(t, err)
		err = session.Advanced().Attachments().StoreEntity(user, names[2], fileStream, "")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var user *User
		err = session.Load(&user, "users/1")
		assert.NoError(t, err)
		metadata, err := session.Advanced().GetMetadataFor(&user)
		assert.NoError(t, err)
		v, ok := metadata.Get(ravendb.Constants_Documents_Metadata_FLAGS)
		assert.True(t, ok)
		vStr := v.(string)
		assert.Equal(t, vStr, "HasAttachments")

		attachmentsI, ok := metadata.Get(ravendb.Constants_Documents_Metadata_ATTACHMENTS)
		assert.True(t, ok)
		attachments := attachmentsI.([]interface{})
		assert.Equal(t, len(attachments), 3)

		sort.Strings(names)
		var gotNames []string
		for _, v := range attachments {

			//TODO: dig deeper into what type metadata.Get() returns. It used to be
			// *ravendb.IMetadataDictionary and is now map[string]interface{}
			//attachment := v.(*ravendb.IMetadataDictionary)
			//name, ok := attachment.Get("Name")

			attachment := v.(map[string]interface{})
			name, ok := attachment["Name"]

			assert.True(t, ok)
			gotNames = append(gotNames, name.(string))
		}
		sort.Strings(gotNames)
		assert.Equal(t, names, gotNames)
		session.Close()
	}
}

func attachmentsSession_throwIfStreamIsUseTwice(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		stream := bytes.NewBuffer([]byte{1, 2, 3})

		user := &User{}
		user.setName("Fitzchak")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		err = session.Advanced().Attachments().StoreEntity(user, "profile", stream, "image/png")
		assert.NoError(t, err)
		err = session.Advanced().Attachments().StoreEntity(user, "other", stream, "")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.Error(t, err)
		_, ok := err.(*ravendb.IllegalStateException)
		assert.True(t, ok)

		session.Close()
	}
}

func attachmentsSession_throwWhenTwoAttachmentsWithTheSameNameInSession(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		stream := bytes.NewBuffer([]byte{1, 2, 3})
		stream2 := bytes.NewBuffer([]byte{1, 2, 3, 4, 5})

		user := &User{}
		user.setName("Fitzchak")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		err = session.Advanced().Attachments().StoreEntity(user, "profile", stream, "image/png")
		assert.NoError(t, err)

		err = session.Advanced().Attachments().StoreEntity(user, "profile", stream2, "")
		assert.Error(t, err)
		_, ok := err.(*ravendb.IllegalStateException)
		assert.True(t, ok)

		session.Close()
	}
}

func attachmentsSession_putDocumentAndAttachmentAndDeleteShouldThrow(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		profileStream := bytes.NewBuffer([]byte{1, 2, 3})
		user := &User{}
		user.setName("Fitzchak")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		err = session.Advanced().Attachments().StoreEntity(user, "profile.png", profileStream, "image/png")
		assert.NoError(t, err)

		err = session.DeleteEntity(user)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.Error(t, err)
		_, ok := err.(*ravendb.IllegalStateException)
		assert.True(t, ok)

		session.Close()
	}
}

func attachmentsSession_deleteAttachments(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user := &User{}
		user.setName("Fitzchak")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		stream1 := bytes.NewBuffer([]byte{1, 2, 3})
		stream2 := bytes.NewBuffer([]byte{1, 2, 3, 4, 5, 6})
		stream3 := bytes.NewBuffer([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})
		stream4 := bytes.NewBuffer([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})

		err = session.Advanced().Attachments().StoreEntity(user, "file1", stream1, "image/png")
		assert.NoError(t, err)
		err = session.Advanced().Attachments().StoreEntity(user, "file2", stream2, "image/png")
		assert.NoError(t, err)
		err = session.Advanced().Attachments().StoreEntity(user, "file3", stream3, "image/png")
		assert.NoError(t, err)
		err = session.Advanced().Attachments().StoreEntity(user, "file4", stream4, "image/png")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var user *User
		err = session.Load(&user, "users/1")
		assert.NoError(t, err)

		// test get attachment by its name
		{
			var attachmentResult *ravendb.CloseableAttachmentResult
			attachmentResult, err = session.Advanced().Attachments().Get("users/1", "file2")
			assert.NoError(t, err)
			name := attachmentResult.GetDetails().Name
			assert.Equal(t, name, "file2")
			attachmentResult.Close()
		}

		err = session.Advanced().Attachments().Delete("users/1", "file2")
		assert.NoError(t, err)
		err = session.Advanced().Attachments().DeleteEntity(user, "file4")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var user *User
		err = session.Load(&user, "users/1")
		assert.NoError(t, err)

		metadata, err := session.Advanced().GetMetadataFor(&user)
		assert.NoError(t, err)

		v, ok := metadata.Get(ravendb.Constants_Documents_Metadata_FLAGS)
		assert.True(t, ok)
		assert.Equal(t, v, "HasAttachments")

		attachmentsI, ok := metadata.Get(ravendb.Constants_Documents_Metadata_ATTACHMENTS)
		assert.True(t, ok)
		attachments := attachmentsI.([]interface{})

		assert.Equal(t, len(attachments), 2)

		{
			result, err := session.Advanced().Attachments().Get("users/1", "file1")
			assert.NoError(t, err)
			r := result.GetData()
			file1Bytes, err := ioutil.ReadAll(r)
			assert.NoError(t, err)

			assert.Equal(t, len(file1Bytes), 3)

			result.Close()
		}
		session.Close()
	}
}

func attachmentsSession_deleteAttachmentsUsingCommand(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user := &User{}
		user.setName("Fitzchak")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		stream1 := bytes.NewBuffer([]byte{1, 2, 3})
		stream2 := bytes.NewBuffer([]byte{1, 2, 3, 4, 5, 6})

		err = session.Advanced().Attachments().StoreEntity(user, "file1", stream1, "image/png")
		assert.NoError(t, err)
		err = session.Advanced().Attachments().StoreEntity(user, "file2", stream2, "image/png")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	op := ravendb.NewDeleteAttachmentOperation("users/1", "file2", nil)
	err = store.Operations().Send(op)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		var user *User
		err = session.Load(&user, "users/1")
		assert.NoError(t, err)

		metadata, err := session.Advanced().GetMetadataFor(&user)
		assert.NoError(t, err)

		v, ok := metadata.Get(ravendb.Constants_Documents_Metadata_FLAGS)
		assert.True(t, ok)
		assert.Equal(t, v, "HasAttachments")

		attachmentsI, ok := metadata.Get(ravendb.Constants_Documents_Metadata_ATTACHMENTS)
		assert.True(t, ok)
		attachments := attachmentsI.([]interface{})
		assert.Equal(t, len(attachments), 1)

		{
			result, err := session.Advanced().Attachments().Get("users/1", "file1")
			assert.NoError(t, err)
			r := result.GetData()
			file1Bytes, err := ioutil.ReadAll(r)
			assert.NoError(t, err)
			assert.Equal(t, len(file1Bytes), 3)

			result.Close()
		}

		session.Close()
	}
}

func attachmentsSession_getAttachmentReleasesResources(t *testing.T, driver *RavenTestDriver) {
	count := 30
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user := &User{}
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	for i := 0; i < count; i++ {
		session := openSessionMust(t, store)

		stream1 := bytes.NewBuffer([]byte{1, 2, 3})
		err = session.Advanced().Attachments().Store("users/1", "file"+strconv.Itoa(i), stream1, "image/png")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	for i := 0; i < count; i++ {
		session := openSessionMust(t, store)
		result, err := session.Advanced().Attachments().Get("users/1", "file"+strconv.Itoa(i))
		assert.NoError(t, err)
		// don't read data as it marks entity as consumed
		result.Close()
		session.Close()
	}
}

func attachmentsSession_deleteDocumentAndThanItsAttachments_ThisIsNoOpButShouldBeSupported(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user := &User{}
		user.setName("Fitzchak")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		stream := bytes.NewBuffer([]byte{1, 2, 3})

		err = session.Advanced().Attachments().StoreEntity(user, "file", stream, "image/png")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var user *User
		err = session.Load(&user, "users/1")
		assert.NoError(t, err)

		err = session.DeleteEntity(user)
		assert.NoError(t, err)
		err = session.Advanced().Attachments().DeleteEntity(user, "file")
		assert.NoError(t, err)
		err = session.Advanced().Attachments().DeleteEntity(user, "file") // this should be no-op
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}
}

func attachmentsSession_deleteDocumentByCommandAndThanItsAttachments_ThisIsNoOpButShouldBeSupported(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setName("Fitzchak")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		stream := bytes.NewBuffer([]byte{1, 2, 3})

		err = session.Advanced().Attachments().StoreEntity(user, "file", stream, "image/png")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		cd := ravendb.NewDeleteCommandData("users/1", nil)
		session.Advanced().Defer(cd)
		err = session.Advanced().Attachments().Delete("users/1", "file")
		assert.NoError(t, err)
		err = session.Advanced().Attachments().Delete("users/1", "file")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}
}

func attachmentsSession_getAttachmentNames(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	names := []string{"profile.png"}

	{
		session := openSessionMust(t, store)

		profileStream := bytes.NewBuffer([]byte{1, 2, 3})

		user := &User{}
		user.setName("Fitzchak")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		err = session.Advanced().Attachments().Store("users/1", names[0], profileStream, "image/png")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var user *User
		err = session.Load(&user, "users/1")
		assert.NoError(t, err)

		attachments, err := session.Advanced().Attachments().GetNames(&user)
		assert.NoError(t, err)

		assert.Equal(t, len(attachments), 1)

		attachment := attachments[0]
		assert.Equal(t, attachment.ContentType, "image/png")
		assert.Equal(t, attachment.Name, names[0])
		assert.Equal(t, attachment.Size, int64(3))

		session.Close()
	}
}

func attachmentsSession_attachmentExists(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		stream := bytes.NewBuffer([]byte{1, 2, 3})

		user := &User{}
		user.setName("Fitzchak")

		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		err = session.Advanced().Attachments().Store("users/1", "profile", stream, "image/png")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		ok, err := session.Advanced().Attachments().Exists("users/1", "profile")
		assert.NoError(t, err)
		assert.True(t, ok)

		ok, err = session.Advanced().Attachments().Exists("users/1", "background-photo")
		assert.NoError(t, err)
		assert.False(t, ok)

		ok, err = session.Advanced().Attachments().Exists("users/2", "profile")
		assert.NoError(t, err)
		assert.False(t, ok)

		session.Close()
	}
}

func TestAttachmentsSession(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	attachmentsSession_putAttachments(t, driver)
	attachmentsSession_putDocumentAndAttachmentAndDeleteShouldThrow(t, driver)
	attachmentsSession_getAttachmentNames(t, driver)

	if enableFlakyTests {
		attachmentsSession_deleteDocumentByCommandAndThanItsAttachments_ThisIsNoOpButShouldBeSupported(t, driver)
	}
	if enableFlakyTests {
		attachmentsSession_deleteAttachments(t, driver)
	}
	if enableFlakyTests {
		attachmentsSession_attachmentExists(t, driver)
	}
	attachmentsSession_throwWhenTwoAttachmentsWithTheSameNameInSession(t, driver)
	if enableFlakyTests {
		attachmentsSession_deleteDocumentAndThanItsAttachments_ThisIsNoOpButShouldBeSupported(t, driver)
	}
	attachmentsSession_throwIfStreamIsUseTwice(t, driver)
	if enableFlakyTests {
		attachmentsSession_getAttachmentReleasesResources(t, driver)
	}
	if enableFlakyTests {
		attachmentsSession_deleteAttachmentsUsingCommand(t, driver)
	}
}
