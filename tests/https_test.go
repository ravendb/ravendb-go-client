package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func httpsTest_canConnectWithCertificate(t *testing.T) {
	var err error
	store := getSecuredDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		user1 := &User{}
		user1.setLastName("user1")
		err = newSession.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
	}
}

func TestHttps(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	// self-signing cert on windows is not added as root ca
	if isWindows() {
		fmt.Printf("Skipping TestHttps on windows\n")
		t.Skip("Skipping on windows")
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of java tests
	httpsTest_canConnectWithCertificate(t)
}
