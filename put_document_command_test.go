package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
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
	}
}

func TestPutDocumentCommand(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_put_document_command_go.txt")
	}

	createTestDriver()
	defer deleteTestDriver()

	putDocumentCOmmand_canPutDocumentUsingCommand(t)
}
