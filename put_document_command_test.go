package ravendb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func putDocumentCOmmand_canPutDocumentUsingCommand(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	user := NewUser()
	user.setName("Marcin")
	user.setAge(30)

	node := valueToTree(user)
	command := NewPutDocumentCommand("users/1", nil, node)
	err = store.GetRequestExecutor().executeCommand(command)
	assert.NoError(t, err)

	result := command.Result
	assert.Equal(t, "users/1", result.getId())

	assert.NotNil(t, result.getChangeVector())

	{
		session := openSessionMust(t, store)
		loadedUserI, err := session.load(getTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		loadedUser := loadedUserI.(*User)
		assert.Equal(t, "Marcin", *loadedUser.getName())
		session.Close()
	}
}

func TestPutDocumentCommand(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered in %s\n", t.Name())
		}
		destroyDriver()
	}()

	putDocumentCOmmand_canPutDocumentUsingCommand(t)
}
