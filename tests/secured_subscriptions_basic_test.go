package tests

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func securedSubscriptionsBasic_shouldStreamAllDocumentsAfterSubscriptionCreation(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getSecuredDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user1 := &User{
			Age: 31,
		}
		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)

		user2 := &User{
			Age: 27,
		}
		err = session.StoreWithID(user2, "users/12")
		assert.NoError(t, err)

		user3 := &User{
			Age: 25,
		}
		err = session.StoreWithID(user3, "users/3")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	id, err := store.Subscriptions.CreateForType(reflect.TypeOf(&User{}), nil, "")
	assert.NoError(t, err)

	{
		opts, err := ravendb.NewSubscriptionWorkerOptions(id)
		assert.NoError(t, err)
		clazz := reflect.TypeOf(&User{})
		subscription, err := store.Subscriptions.GetSubscriptionWorker(clazz, opts, "")
		assert.NoError(t, err)

		keys := make(chan string)
		ages := make(chan int)

		fn := func(batch *ravendb.SubscriptionBatch) error {
			// Note: important that done in two separate passes
			for _, item := range batch.Items {
				keys <- item.ID
			}

			for _, item := range batch.Items {
				v, err := item.GetResult()
				assert.NoError(t, err)
				u := v.(*User)
				ages <- u.Age
			}
			return nil
		}
		_, err = subscription.Run(fn)
		assert.NoError(t, err)

		getNextKey := func() string {
			select {
			case v := <-keys:
				return v
			case <-time.After(_reasonableWaitTime):
				// no-op
			}
			return ""
		}
		key := getNextKey()
		assert.Equal(t, key, "users/1")
		key = getNextKey()
		assert.Equal(t, key, "users/12")
		key = getNextKey()
		assert.Equal(t, key, "users/3")

		getNextAge := func() int {
			select {
			case v := <-ages:
				return v
			case <-time.After(_reasonableWaitTime):
				// no-op
			}
			return 0
		}
		age := getNextAge()
		assert.Equal(t, age, 31)
		age = getNextAge()
		assert.Equal(t, age, 27)
		age = getNextAge()
		assert.Equal(t, age, 25)

		err = subscription.Close()
		assert.NoError(t, err)
	}

}

func securedSubscriptionsBasic_shouldSendAllNewAndModifiedDocs(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getSecuredDocumentStoreMust(t)
	defer store.Close()

	id, err := store.Subscriptions.CreateForType(reflect.TypeOf(&User{}), nil, "")
	assert.NoError(t, err)

	{
		opts, err := ravendb.NewSubscriptionWorkerOptions(id)
		assert.NoError(t, err)
		clazz := reflect.TypeOf(map[string]interface{}{})
		subscription, err := store.Subscriptions.GetSubscriptionWorker(clazz, opts, "")
		assert.NoError(t, err)

		names := make(chan string, 20)

		processBatch := func(batch *ravendb.SubscriptionBatch) error {
			for _, item := range batch.Items {
				v, err := item.GetResult()
				assert.NoError(t, err)
				m := v.(map[string]interface{})
				name := m["name"].(string)
				names <- name
			}
			return nil
		}

		{
			session := openSessionMust(t, store)

			user := &User{}
			user.setName("James")
			err = session.StoreWithID(user, "users/1")
			assert.NoError(t, err)

			err = session.SaveChanges()
			assert.NoError(t, err)

			session.Close()
		}

		_, err = subscription.Run(processBatch)
		assert.NoError(t, err)

		getNextName := func() string {
			select {
			case v := <-names:
				return v
			case <-time.After(_reasonableWaitTime):
				// no-op
			}
			return ""
		}

		name := getNextName()
		assert.Equal(t, name, "James")

		{
			session := openSessionMust(t, store)

			user := &User{}
			user.setName("Adam")
			err = session.StoreWithID(user, "users/12")
			assert.NoError(t, err)

			err = session.SaveChanges()
			assert.NoError(t, err)

			session.Close()
		}

		name = getNextName()
		assert.Equal(t, name, "Adam")

		//Thread.sleep(15000); // test with sleep - let few heartbeats come to us - commented out for CI

		{
			session := openSessionMust(t, store)

			user := &User{}
			user.setName("David")
			err = session.StoreWithID(user, "users/1")
			assert.NoError(t, err)

			err = session.SaveChanges()
			assert.NoError(t, err)

			session.Close()
		}

		name = getNextName()
		assert.Equal(t, name, "David")

		err = subscription.Close()
		assert.NoError(t, err)
	}
}

func TestSecuredSubscriptionsBasic(t *testing.T) {
	// t.Parallel()

	// self-signing cert on windows is not added as root ca
	if isWindows() {
		fmt.Printf("Skipping TestHttps on windows\n")
		t.Skip("Skipping on windows")
		return
	}

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	securedSubscriptionsBasic_shouldStreamAllDocumentsAfterSubscriptionCreation(t, driver)
	securedSubscriptionsBasic_shouldSendAllNewAndModifiedDocs(t, driver)
}
