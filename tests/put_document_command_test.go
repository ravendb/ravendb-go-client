package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func putDocumentCOmmand_canPutDocumentUsingCommand(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	user := NewUser()
	user.setName("Marcin")
	user.setAge(30)

	node := ravendb.ValueToTree(user)
	command := ravendb.NewPutDocumentCommand("users/1", nil, node)
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)

	result := command.Result
	assert.Equal(t, "users/1", result.GetID())

	assert.NotNil(t, result.GetChangeVector())

	{
		session := openSessionMust(t, store)
		loadedUserI, err := session.Load(ravendb.GetTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		loadedUser := loadedUserI.(*User)
		assert.Equal(t, "Marcin", *loadedUser.GetName())
		session.Close()
	}
}

func TestPutDocumentCommand(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	putDocumentCOmmand_canPutDocumentUsingCommand(t)
}
