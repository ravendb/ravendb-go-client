package tests

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func revisionsSubscriptions_plainRevisionsSubscriptions(t *testing.T, driver *RavenTestDriver) {
	if isRunningOn41Server() {
		return
	}

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	subscriptionId, err := store.Subscriptions.CreateForRevisions(reflect.TypeOf(&User{}), nil, "")
	assert.NoError(t, err)

	defaultCollection := &ravendb.RevisionsCollectionConfiguration{
		MinimumRevisionsToKeep: 5,
	}

	usersConfig := &ravendb.RevisionsCollectionConfiguration{}
	donsConfig := &ravendb.RevisionsCollectionConfiguration{}

	configuration := &ravendb.RevisionsConfiguration{
		DefaultConfig: defaultCollection,
	}
	perCollectionConfig := map[string]*ravendb.RevisionsCollectionConfiguration{
		"Users": usersConfig,
		"Dons":  donsConfig,
	}

	configuration.Collections = perCollectionConfig

	operation := ravendb.NewConfigureRevisionsOperation(configuration)

	err = store.Maintenance().Send(operation)
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			{
				session, err := store.OpenSession()
				assert.NoError(t, err)

				user := &User{}
				user.setName(fmt.Sprintf("users%d ver %d", i, j))
				err = session.StoreWithID(user, "users/"+strconv.Itoa(i))
				assert.NoError(t, err)

				company := &Company{
					Name: fmt.Sprintf("dons%d ver %d", i, j),
				}
				err = session.StoreWithID(company, "dons/"+strconv.Itoa(i))
				assert.NoError(t, err)

				err = session.SaveChanges()
				assert.NoError(t, err)

				session.Close()
			}
		}
	}

	{
		opts, err := ravendb.NewSubscriptionWorkerOptions(subscriptionId)
		assert.NoError(t, err)
		clazz := reflect.TypeOf(&User{})
		sub, err := store.Subscriptions.GetSubscriptionWorkerForRevisions(clazz, opts, "")
		assert.NoError(t, err)

		mre := make(chan bool)
		names := map[string]struct{}{}

		fn := func(x *ravendb.SubscriptionBatch) error {
			for _, item := range x.Items {
				// result is ravendb.Revision of type User
				v, err := item.GetResult()
				assert.NoError(t, err)
				result := v.(*ravendb.Revision)
				var currName string
				var prevName string
				if result.Current != nil {
					u := result.Current.(*User)
					currName = *u.Name
				}
				if result.Previous != nil {
					u := result.Previous.(*User)
					prevName = *u.Name
				}
				name := currName + prevName
				names[name] = struct{}{}
				if len(names) == 100 {
					mre <- true
				}
			}
			return nil
		}
		_, err = sub.Run(fn)
		assert.NoError(t, err)

		timedOut := chanWaitTimedOut(mre, _reasonableWaitTime)
		assert.False(t, timedOut)

		err = sub.Close()
		assert.NoError(t, err)
	}
}

func revisionsSubscriptions_plainRevisionsSubscriptionsCompareDocs(t *testing.T, driver *RavenTestDriver) {
	if isRunningOn41Server() {
		return
	}

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	subscriptionId, err := store.Subscriptions.CreateForRevisions(reflect.TypeOf(&User{}), nil, "")
	assert.NoError(t, err)

	defaultCollection := &ravendb.RevisionsCollectionConfiguration{
		MinimumRevisionsToKeep: 5,
	}

	usersConfig := &ravendb.RevisionsCollectionConfiguration{}
	donsConfig := &ravendb.RevisionsCollectionConfiguration{}

	configuration := &ravendb.RevisionsConfiguration{
		DefaultConfig: defaultCollection,
	}
	perCollectionConfig := map[string]*ravendb.RevisionsCollectionConfiguration{
		"Users": usersConfig,
		"Dons":  donsConfig,
	}

	configuration.Collections = perCollectionConfig

	operation := ravendb.NewConfigureRevisionsOperation(configuration)

	err = store.Maintenance().Send(operation)
	assert.NoError(t, err)

	for j := 0; j < 10; j++ {
		{
			session, err := store.OpenSession()
			assert.NoError(t, err)

			user := &User{
				Age: j,
			}
			user.setName("users1 ver " + strconv.Itoa(j))
			err = session.StoreWithID(user, "users/1")
			assert.NoError(t, err)

			company := &Company{
				Name: "dons1 ver " + strconv.Itoa(j),
			}
			err = session.StoreWithID(company, "dons/1")
			assert.NoError(t, err)

			err = session.SaveChanges()
			assert.NoError(t, err)

			session.Close()
		}
	}

	{
		opts, err := ravendb.NewSubscriptionWorkerOptions(subscriptionId)
		assert.NoError(t, err)
		clazz := reflect.TypeOf(&User{})
		sub, err := store.Subscriptions.GetSubscriptionWorkerForRevisions(clazz, opts, "")
		assert.NoError(t, err)

		mre := make(chan bool)
		names := map[string]struct{}{}

		var mu sync.Mutex
		maxAge := -1

		fn := func(a *ravendb.SubscriptionBatch) error {
			for _, item := range a.Items {
				// result is ravendb.Revision of type User
				v, err := item.GetResult()
				assert.NoError(t, err)
				result := v.(*ravendb.Revision)

				var currName, prevName string
				currAge := -1
				prevAge := -1
				if result.Current != nil {
					u := result.Current.(*User)
					currName = *u.Name
					currAge = u.Age
				}
				if result.Previous != nil {
					u := result.Previous.(*User)
					prevName = *u.Name
					prevAge = u.Age
				}

				mu.Lock()

				if currAge > maxAge && currAge > prevAge {
					name := currName + " " + prevName
					names[name] = struct{}{}
					maxAge = currAge
				}

				shouldRelease := len(names) == 10
				mu.Unlock()

				if shouldRelease {
					mre <- true
				}
			}
			return nil
		}
		_, err = sub.Run(fn)
		assert.NoError(t, err)

		timedOut := chanWaitTimedOut(mre, _reasonableWaitTime)
		assert.False(t, timedOut)

		err = sub.Close()
		assert.NoError(t, err)
	}
}

func TestRevisionsSubscriptions(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	revisionsSubscriptions_plainRevisionsSubscriptionsCompareDocs(t, driver)
	revisionsSubscriptions_plainRevisionsSubscriptions(t, driver)
}
