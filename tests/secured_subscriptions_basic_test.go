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

	var users []*User
	{
		session := openSessionMust(t, store)

		user1 := &User{
			Age: 31,
		}
		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		users = append(users, user1)

		user2 := &User{
			Age: 27,
		}
		err = session.StoreWithID(user2, "users/12")
		assert.NoError(t, err)
		users = append(users, user2)

		user3 := &User{
			Age: 25,
		}
		err = session.StoreWithID(user3, "users/3")
		assert.NoError(t, err)
		users = append(users, user3)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	id, err := store.Subscriptions().CreateForType(reflect.TypeOf(&User{}), nil, "")
	assert.NoError(t, err)

	{
		opts := ravendb.NewSubscriptionWorkerOptions(id)
		clazz := reflect.TypeOf(&User{})
		subscription, err := store.Subscriptions().GetSubscriptionWorker(clazz, opts, "")
		assert.NoError(t, err)

		results, err := subscription.Run()
		assert.NoError(t, err)
		select {
			case batch := <- results:
				for i, item := range batch.Items {
					userExp := users[i]
					assert.Equal(t, item.ID, userExp.ID)
					v, err := item.GetResult()
					assert.NoError(t, err)
					u := v.(*User)
					assert.Equal(t, u.Age, userExp.Age)
				}

				case <- time.After(_reasonableWaitTime):
					assert.Fail(t, "timed out waiting for batch")
		}
		err = subscription.Close()
		assert.NoError(t, err)
	}

}

func securedSubscriptionsBasic_shouldSendAllNewAndModifiedDocs(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getSecuredDocumentStoreMust(t)
	defer store.Close()

	id, err := store.Subscriptions().CreateForType(reflect.TypeOf(&User{}), nil, "")
	assert.NoError(t, err)

	{
		opts := ravendb.NewSubscriptionWorkerOptions(id)
		clazz := reflect.TypeOf(map[string]interface{}{})
		subscription, err := store.Subscriptions().GetSubscriptionWorker(clazz, opts, "")
		assert.NoError(t, err)

		names := make(chan string, 20)

		results, err := subscription.Run()
		assert.NoError(t, err)

		processBatch := func(batch *ravendb.SubscriptionBatch) {
			for _, item := range batch.Items {
				v, err := item.GetResult()
				assert.NoError(t, err)
				m := v.(map[string]interface{})
				name := m["name"].(string)
				names <- name
			}
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

		select {
		case batch := <-results:
			processBatch((batch))
			case <- time.After(_reasonableWaitTime):
				assert.Fail(t, "failed waiting for batch")
		}

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
