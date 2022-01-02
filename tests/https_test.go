package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func httpsTestCanConnectWithCertificate(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getSecuredDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store) //may need to run the IDE from administrator
		user1 := &User{}
		user1.setLastName("user1")
		err = newSession.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
	}
}

func TestHttps(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of java tests
	httpsTestCanConnectWithCertificate(t, driver)
}
