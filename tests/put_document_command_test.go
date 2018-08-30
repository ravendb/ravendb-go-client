package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func putDocumentCOmmand_canPutDocumentUsingCommand(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	user := &User{}
	user.setName("Marcin")
	user.Age = 30

	node := ravendb.ValueToTree(user)
	command := ravendb.NewPutDocumentCommand("users/1", nil, node)
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)

	result := command.Result
	assert.Equal(t, "users/1", result.GetID())

	assert.NotNil(t, result.GetChangeVector())

	{
		session := openSessionMust(t, store)
		loadedUserI, err := session.LoadOld(ravendb.GetTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		loadedUser := loadedUserI.(*User)
		assert.Equal(t, "Marcin", *loadedUser.Name)
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
