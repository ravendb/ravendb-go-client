package tests

import (
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

const (
	_reasonableWaitTime = time.Second * 5 // TODO: is it 60 seconds in Java?
)

// returns true if timed out
func chanWaitTimedOut(ch chan bool, timeout time.Duration) bool {
	select {
	case <-ch:
		return false
	case <-time.After(timeout):
		return true
	}
}

// returns false if timed out
func getNextUser(docs chan *User, timeout time.Duration) (*User, bool) {
	if timeout == 0 {
		timeout = _reasonableWaitTime
	}
	select {
	case u := <-docs:
		return u, true
	case <-time.After(timeout):
		return nil, false
	}
}

func subscriptionsBasic_canDeleteSubscription(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	id1, err := store.Subscriptions().CreateForType(reflect.TypeOf(&User{}), nil, "")
	assert.NoError(t, err)
	id2, err := store.Subscriptions().CreateForType(reflect.TypeOf(&User{}), nil, "")
	assert.NoError(t, err)

	subscriptions, err := store.Subscriptions().GetSubscriptions(0, 5, "")
	assert.NoError(t, err)
	assert.Equal(t, len(subscriptions), 2)

	// test getSubscriptionState as well
	subscriptionState, err := store.Subscriptions().GetSubscriptionState(id1, "")
	assert.NoError(t, err)
	cv := subscriptionState.ChangeVectorForNextBatchStartingPoint
	assert.Nil(t, cv)

	err = store.Subscriptions().Delete(id1, "")
	assert.NoError(t, err)
	err = store.Subscriptions().Delete(id2, "")
	assert.NoError(t, err)

	subscriptions, err = store.Subscriptions().GetSubscriptions(0, 5, "")
	assert.NoError(t, err)
	assert.Equal(t, len(subscriptions), 0)
}

func subscriptionsBasic_shouldThrowWhenOpeningNoExistingSubscription(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	clazz := reflect.TypeOf(&map[string]interface{}{})
	opts := ravendb.NewSubscriptionWorkerOptions("1")

	subscription, err := store.Subscriptions().GetSubscriptionWorker(clazz, opts, "")
	assert.NoError(t, err)

	_, err = subscription.Run()
	assert.NoError(t, err)
	err = subscription.WaitUntilFinished(0)
	assert.NotNil(t, err)
	_, ok := err.(*ravendb.SubscriptionDoesNotExistError)
	assert.True(t, ok)
	assert.Equal(t, err, subscription.Err())

	err = subscription.Close()
	assert.NoError(t, err)
}

func subscriptionsBasic_shouldThrowOnAttemptToOpenAlreadyOpenedSubscription(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	id, err := store.Subscriptions().CreateForType(reflect.TypeOf(&User{}), nil, "")
	assert.NoError(t, err)

	{
		clazz := reflect.TypeOf(map[string]interface{}{})
		opts := ravendb.NewSubscriptionWorkerOptions(id)
		subscription, err := store.Subscriptions().GetSubscriptionWorker(clazz, opts, "")
		assert.NoError(t, err)

		{
			session, err := store.OpenSession("")
			assert.NoError(t, err)
			err = session.Store(&User{})
			assert.NoError(t, err)
			err = session.SaveChanges()
			assert.NoError(t, err)

			session.Close()
		}

		results, err := subscription.Run()
		assert.NoError(t, err)

		select {
			case <-results:
				// no-op, got the result
				case <-time.After(_reasonableWaitTime):
					// no-op, timeout waiting for the result
		}

		options2 := ravendb.NewSubscriptionWorkerOptions(id)
		options2.Strategy = ravendb.SubscriptionOpeningStrategyOpenIfFree

		{
			secondSubscription, err := store.Subscriptions().GetSubscriptionWorker(clazz, options2, "")
			assert.NoError(t, err)
			_, err = secondSubscription.Run()
			assert.NoError(t, err)
			err = secondSubscription.WaitUntilFinished(0)
			_, ok := err.(*ravendb.SubscriptionInUseError)
			assert.True(t, ok)

			err = secondSubscription.Close()
			assert.NoError(t, err)
		}

		err = subscription.Close()
		assert.NoError(t, err)
	}

}

func subscriptionsBasic_shouldStreamAllDocumentsAfterSubscriptionCreation(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
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

		chDone := make(chan bool, 1)
		go func() {
			for items := range results {
				for i, item := range items {
					expUser := users[i]
					assert.Equal(t, item.ID, expUser.ID)
					v, err := item.GetResult()
					assert.NoError(t, err)
					u := v.(*User)
					assert.Equal(t, u.Age, expUser.Age)
				}
				chDone <- true
			}
		}()
		select {
		case <-chDone:
			// no-op, got the first batch
			case <-time.After(_reasonableWaitTime):
				assert.False(t, true, "timed out waiting for batch")
		}

		err = subscription.Close()
		assert.NoError(t, err)
	}
}

func subscriptionsBasic_shouldSendAllNewAndModifiedDocs(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	id, err := store.Subscriptions().CreateForType(reflect.TypeOf(&User{}), nil, "")
	assert.NoError(t, err)

	{
		opts := ravendb.NewSubscriptionWorkerOptions(id)
		clazz := reflect.TypeOf(map[string]interface{}{})
		subscription, err := store.Subscriptions().GetSubscriptionWorker(clazz, opts, "")
		assert.NoError(t, err)

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

		results, err := subscription.Run()
		assert.NoError(t, err)
		go func() {
			expNames := []string{"James", "Adam", "David"}
			var n int
			for items := range results {
				for _, item := range items {
					v, err := item.GetResult()
					assert.NoError(t, err)
					m := v.(map[string]interface{})
					name := m["name"].(string)
					assert.Equal(t, name, expNames[n])
					n++
				}
			}
		}()

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

		err = subscription.Close()
		assert.NoError(t, err)
	}
}

func subscriptionsBasic_shouldRespectMaxDocCountInBatch(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		for i := 0; i < 100; i++ {
			err = session.Store(&Company{})
			assert.NoError(t, err)
		}

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	clazz := reflect.TypeOf(&Company{})
	id, err := store.Subscriptions().CreateForType(clazz, nil, "")
	assert.NoError(t, err)

	options := ravendb.NewSubscriptionWorkerOptions(id)
	options.MaxDocsPerBatch = 25

	{
		clazz = reflect.TypeOf(map[string]interface{}{})
		subscriptionWorker, err := store.Subscriptions().GetSubscriptionWorker(clazz, options, "")
		assert.NoError(t, err)

		results, err := subscriptionWorker.Run()
		assert.NoError(t, err)

		var totalItems int
		for totalItems < 100 {
			select {
			case  items := <-results:
				n := len(items)
				assert.True(t, n <= 25)
				totalItems += n
			case <-time.After(_reasonableWaitTime):
				assert.False(t, true, "timed out waiting for a batch")
				totalItems = 100
			}
		}
		_ = subscriptionWorker.Close()
	}
}

func subscriptionsBasic_shouldRespectCollectionCriteria(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		for i := 0; i < 100; i++ {
			err = session.Store(&Company{})
			assert.NoError(t, err)
			err = session.Store(&User{})
			assert.NoError(t, err)
		}

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	clazz := reflect.TypeOf(&User{})
	id, err := store.Subscriptions().CreateForType(clazz, nil, "")
	assert.NoError(t, err)

	options := ravendb.NewSubscriptionWorkerOptions(id)
	options.MaxDocsPerBatch = 31

	{
		clazz = reflect.TypeOf(map[string]interface{}{})
		subscriptionWorker, err := store.Subscriptions().GetSubscriptionWorker(clazz, options, "")
		assert.NoError(t, err)

		results, err := subscriptionWorker.Run()
		assert.NoError(t, err)

		var totalItems int
		for totalItems < 100 {
			select {
			case items := <-results:
				n := len(items)
				assert.True(t, n <= 31)
				totalItems += n
				case <-time.After(_reasonableWaitTime):
					assert.Fail(t, "timed out waiting for batch")
					totalItems = 100
			}

		}

		_ = subscriptionWorker.Close()
	}
}

func subscriptionsBasic_willAcknowledgeEmptyBatches(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	subscriptionDocuments, err := store.Subscriptions().GetSubscriptions(0, 10, "")
	assert.NoError(t, err)
	assert.Equal(t, len(subscriptionDocuments), 0)

	opts := &ravendb.SubscriptionCreationOptions{}
	clazz := reflect.TypeOf(&User{})
	allId, err := store.Subscriptions().CreateForType(clazz, opts, "")
	assert.NoError(t, err)

	{
		clazz = reflect.TypeOf(map[string]interface{}{})
		opts := ravendb.NewSubscriptionWorkerOptions(allId)
		allSubscription, err := store.Subscriptions().GetSubscriptionWorker(clazz, opts, "")
		assert.NoError(t, err)

		var allCounter int32
		allSemaphore := make(chan bool)

		filteredOptions := &ravendb.SubscriptionCreationOptions{
			Query: "from Users where age < 0",
		}
		filteredUsersId, err := store.Subscriptions().Create(filteredOptions, "")
		assert.NoError(t, err)

		{
			clazz = reflect.TypeOf(map[string]interface{}{})
			opts := ravendb.NewSubscriptionWorkerOptions(filteredUsersId)
			filteredUsersSubscription, err := store.Subscriptions().GetSubscriptionWorker(clazz, opts, "")
			assert.NoError(t, err)

			usersDocsSemaphore := make(chan bool)

			{
				session := openSessionMust(t, store)

				for i := 0; i < 500; i++ {
					err = session.StoreWithID(&User{}, "another/")
					assert.NoError(t, err)
				}

				err = session.SaveChanges()
				assert.NoError(t, err)

				session.Close()
			}

			results, err := allSubscription.Run()
			assert.NoError(t, err)

			go func() {
				for items := range results {
					n := len(items)
					total := atomic.AddInt32(&allCounter, int32(n))
					if total >= 100 {
						allSemaphore <- true
					}
				}
			}()

			results2, err := filteredUsersSubscription.Run()
			assert.NoError(t, err)

			// TODO: more go-ish waiting using select on 2 channels
			go func() {
				for range results2 {
					usersDocsSemaphore <- true
				}
			}()

			timedOut := chanWaitTimedOut(allSemaphore, _reasonableWaitTime)
			assert.False(t, timedOut)
			timedOut = chanWaitTimedOut(usersDocsSemaphore, time.Millisecond*50)
			assert.True(t, timedOut)
		}

		_ = allSubscription.Close()
	}
}

func subscriptionsBasic_canReleaseSubscription(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	var subscriptionWorker *ravendb.SubscriptionWorker
	var throwingSubscriptionWorker *ravendb.SubscriptionWorker
	var notThrowingSubscriptionWorker *ravendb.SubscriptionWorker

	defer func() {
		if subscriptionWorker != nil {
			_ = subscriptionWorker.Close()
		}
		if throwingSubscriptionWorker != nil {
			_ = throwingSubscriptionWorker.Close()
		}
		if notThrowingSubscriptionWorker != nil {
			_ = notThrowingSubscriptionWorker.Close()
		}
	}()

	opts := &ravendb.SubscriptionCreationOptions{}
	clazz := reflect.TypeOf(&User{})
	id, err := store.Subscriptions().CreateForType(clazz, opts, "")
	assert.NoError(t, err)

	options1 := ravendb.NewSubscriptionWorkerOptions(id)
	options1.Strategy = ravendb.SubscriptionOpeningStrategyOpenIfFree
	clazz = reflect.TypeOf(map[string]interface{}{})
	subscriptionWorker, err = store.Subscriptions().GetSubscriptionWorker(clazz, options1, "")
	assert.NoError(t, err)

	putUserDoc(t, store)

	results, err := subscriptionWorker.Run()
	assert.NoError(t, err)
	select {
		case <-results:
			// no-op, got a result
			case <-time.After(_reasonableWaitTime):
				assert.Fail(t, "timed out waiting for batch")
	}

	options2 := ravendb.NewSubscriptionWorkerOptions(id)
	options2.Strategy = ravendb.SubscriptionOpeningStrategyOpenIfFree
	throwingSubscriptionWorker, err = store.Subscriptions().GetSubscriptionWorker(clazz, options2, "")
	assert.NoError(t, err)

	_, err = throwingSubscriptionWorker.Run()
	err = throwingSubscriptionWorker.WaitUntilFinished(0)
	_, ok := err.(*ravendb.SubscriptionInUseError)
	assert.True(t, ok)

	err = store.Subscriptions().DropConnection(id, "")
	assert.NoError(t, err)

	wopts := ravendb.NewSubscriptionWorkerOptions(id)
	notThrowingSubscriptionWorker, err = store.Subscriptions().GetSubscriptionWorker(clazz, wopts, "")
	results, err = notThrowingSubscriptionWorker.Run()
	assert.NoError(t, err)
	putUserDoc(t, store)

	select {
	case <-results:
	// no-op, got a result
	case <-time.After(_reasonableWaitTime):
		assert.Fail(t, "timed out waiting for batch")
	}
}

func putUserDoc(t *testing.T, store *ravendb.DocumentStore) {
	session, err := store.OpenSession("")
	assert.NoError(t, err)
	defer session.Close()

	err = session.Store(&User{})
	assert.NoError(t, err)
	err = session.SaveChanges()
	assert.NoError(t, err)
}

func subscriptionsBasic_shouldPullDocumentsAfterBulkInsert(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	opts := &ravendb.SubscriptionCreationOptions{}
	id, err := store.Subscriptions().CreateForType(reflect.TypeOf(&User{}), opts, "")
	assert.NoError(t, err)

	{
		clazz := reflect.TypeOf(&User{})
		wopts := ravendb.NewSubscriptionWorkerOptions(id)
		subscription, err := store.Subscriptions().GetSubscriptionWorker(clazz, wopts, "")
		{
			bulk := store.BulkInsert("")
			_, err = bulk.Store(&User{}, nil)
			assert.NoError(t, err)
			_, err = bulk.Store(&User{}, nil)
			assert.NoError(t, err)
			_, err = bulk.Store(&User{}, nil)
			assert.NoError(t, err)
			err = bulk.Close()
			assert.NoError(t, err)
		}

		results, err := subscription.Run()
		done := false
		nUsers := 0
		for !done {
			select {
			case items := <- results:
				for _, item := range items {
					v, err := item.GetResult()
					assert.NoError(t, err)
					u := v.(*User)
					assert.NotNil(t, u)
					nUsers++
					if nUsers >= 2 {
						done = true
					}
				}
				case <-time.After(_reasonableWaitTime):
					done = true
					assert.Fail(t, "timed out waiting for batch")
			}
		}

		err = subscription.Close()
		assert.NoError(t, err)
	}
}

func subscriptionsBasic_shouldStopPullingDocsAndCloseSubscriptionOnSubscriberErrorByDefault(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	opts := &ravendb.SubscriptionCreationOptions{}
	id, err := store.Subscriptions().CreateForType(reflect.TypeOf(&User{}), opts, "")
	assert.NoError(t, err)

	{
		clazz := reflect.TypeOf(map[string]interface{}{})
		wopts := ravendb.NewSubscriptionWorkerOptions(id)
		subscription, err := store.Subscriptions().GetSubscriptionWorker(clazz, wopts, "")

		putUserDoc(t, store)

		_, err = subscription.Run()
		assert.NoError(t, err)

		err = subscription.WaitUntilFinished(_reasonableWaitTime)
		_, ok := err.(*ravendb.SubscriberErrorError)
		assert.True(t, ok)
		assert.NotNil(t, subscription.Err())

		res, err := store.Subscriptions().GetSubscriptions(0, 1, "")
		assert.NoError(t, err)
		subscriptionConfig := res[0]
		assert.Nil(t, subscriptionConfig.ChangeVectorForNextBatchStartingPoint)

		err = subscription.Close()
		assert.NoError(t, err)
	}
}


// Note: not applicable in Go becaues batch processing cannot affect
// subscription worker
/*
func subscriptionsBasic_canSetToIgnoreSubscriberErrors(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	opts := &ravendb.SubscriptionCreationOptions{}
	id, err := store.Subscriptions().CreateForType(reflect.TypeOf(&User{}), opts, "")
	assert.NoError(t, err)

	options1 := ravendb.NewSubscriptionWorkerOptions(id)
	options1.IgnoreSubscriberErrors = true

	{
		clazz := reflect.TypeOf(&User{})
		subscription, err := store.Subscriptions().GetSubscriptionWorker(clazz, options1, "")
		assert.NoError(t, err)

		putUserDoc(t, store)
		putUserDoc(t, store)

		results, err := subscription.Run()
		assert.NoError(t, err)

		done := false
		nUsers := 0
		for !done {
			select {
			case batch := <- results:
				for _, item := range batch.Items {
					v, err := item.GetResult()
					assert.NoError(t, err)
					u := v.(*User)
					assert.NotNil(t, u)
					nUsers++
					if nUsers >= 2 {
						done = true
					}
				}
				case <- time.After(_reasonableWaitTime):
					assert.Fail(t, "timed out waiting for batch")
					done = true
			}
		}
		_ = subscription.Close()
		err = subscription.WaitUntilFinished(0)
		assert.NoError(t, err)

		// nno error because we asked to ignore errors
		assert.NoError(t, subscription.Err())

		err = subscription.Close()
		assert.NoError(t, err)
	}
}
*/

func subscriptionsBasic_ravenDB_3452_ShouldStopPullingDocsIfReleased(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	opts := &ravendb.SubscriptionCreationOptions{}
	id, err := store.Subscriptions().CreateForType(reflect.TypeOf(&User{}), opts, "")
	assert.NoError(t, err)

	{
		options1 := ravendb.NewSubscriptionWorkerOptions(id)
		options1.TimeToWaitBeforeConnectionRetry = ravendb.Duration(time.Second)

		clazz := reflect.TypeOf(&User{})
		subscription, err := store.Subscriptions().GetSubscriptionWorker(clazz, options1, "")
		assert.NoError(t, err)

		{
			session, err := store.OpenSession("")
			assert.NoError(t, err)
			err = session.StoreWithID(&User{}, "users/1")
			assert.NoError(t, err)
			err = session.StoreWithID(&User{}, "users/2")
			assert.NoError(t, err)
			err = session.SaveChanges()
			assert.NoError(t, err)

			session.Close()
		}

		results, err := subscription.Run()
		assert.NoError(t, err)

		done := false
		nUsers := 0
		for !done {
			select {
			case items := <-results:
				for _, item := range items {
					v, err := item.GetResult()
					assert.NoError(t, err)
					u := v.(*User)
					assert.NotNil(t, u)
					nUsers++
					if nUsers >= 2 {
						done = true
					}
				}
				case <- time.After(_reasonableWaitTime):
					assert.Fail(t, "timed out waiting for batch")
					done = true
			}
		}

		err = store.Subscriptions().DropConnection(id, "")
		assert.NoError(t, err)

		// this can exit normally or throw on drop connection
		// depending on exactly where the drop happens
		err = subscription.WaitUntilFinished(_reasonableWaitTime)
		if err != nil {
			_, ok := err.(*ravendb.SubscriptionClosedError)
			assert.True(t, ok)
		}

		{
			session, err := store.OpenSession("")
			assert.NoError(t, err)
			err = session.StoreWithID(&User{}, "users/3")
			assert.NoError(t, err)
			err = session.StoreWithID(&User{}, "users/4")
			assert.NoError(t, err)
			err = session.SaveChanges()
			assert.NoError(t, err)

			session.Close()
		}

		// should get no results since we dropped the connection

		select {
		case batch := <-results:
			// if we get it, it should be nil because we receive
			// from closed channel
			assert.Nil(t, batch)
			case <-time.After(50*time.Millisecond):
				// no-op, timeing out is also valid
		}

		assert.True(t, subscription.IsDone())

		err = subscription.Close()
		assert.NoError(t, err)
	}
}

func subscriptionsBasic_ravenDB_3453_ShouldDeserializeTheWholeDocumentsAfterTypedSubscription(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	opts := &ravendb.SubscriptionCreationOptions{}
	id, err := store.Subscriptions().CreateForType(reflect.TypeOf(&User{}), opts, "")
	assert.NoError(t, err)

	{
		clazz := reflect.TypeOf(&User{})
		wopts := ravendb.NewSubscriptionWorkerOptions(id)

		subscription, err := store.Subscriptions().GetSubscriptionWorker(clazz, wopts, "")
		assert.NoError(t, err)

		var users []*User
		{
			session, err := store.OpenSession("")
			assert.NoError(t, err)
			u := &User{Age: 31}
			err = session.StoreWithID(u, "users/1")
			users = append(users, u)
			assert.NoError(t, err)
			u = &User{Age: 27}
			err = session.StoreWithID(u, "users/12")
			users = append(users, u)
			assert.NoError(t, err)
			u = &User{Age: 25}
			err = session.StoreWithID(u, "users/3")
			users = append(users, u)
			assert.NoError(t, err)
			err = session.SaveChanges()
			assert.NoError(t, err)

			session.Close()
		}

		results, err := subscription.Run()
		assert.NoError(t, err)

		n := 0
		done := false
		for !done {
			select {
			case items := <- results:
				for _, item := range items {
					v, err := item.GetResult()
					assert.NoError(t, err)
					u := v.(*User)
					expU := users[n]
					assert.Equal(t, u.ID, expU.ID)
					assert.Equal(t, u.Age, expU.Age)
					n++
					if n >= len(users) {
						done = true
					}
				}

				case <-time.After(_reasonableWaitTime):
					done = true
					assert.Fail(t, "timed out waiting for batch")
				}

				err = subscription.Close()
				assert.NoError(t, err)
			}
	}
}

func subscriptionsBasic_disposingOneSubscriptionShouldNotAffectOnNotificationsOfOthers(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	var subscription1 *ravendb.SubscriptionWorker
	var subscription2 *ravendb.SubscriptionWorker
	defer func() {
		if subscription1 != nil {
			_ = subscription1.Close()
		}
		if subscription2 != nil {
			_ = subscription2.Close()
		}
	}()

	id1, err := store.Subscriptions().CreateForType(reflect.TypeOf(&User{}), nil, "")
	assert.NoError(t, err)
	id2, err := store.Subscriptions().CreateForType(reflect.TypeOf(&User{}), nil, "")
	assert.NoError(t, err)

	{
		session, err := store.OpenSession("")
		assert.NoError(t, err)
		err = session.StoreWithID(&User{}, "users/1")
		assert.NoError(t, err)
		err = session.StoreWithID(&User{}, "users/2")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	clazz := reflect.TypeOf(&User{})
	opts := ravendb.NewSubscriptionWorkerOptions(id1)

	subscription1, err = store.Subscriptions().GetSubscriptionWorker(clazz, opts, "")
	assert.NoError(t, err)

	items1 := make(chan *User, 10)

	results, err := subscription1.Run()
	assert.NoError(t, err)

	// TODO: rewrite in a more go-ish way
	go func() {
		for items := range results {
			for _, item := range items {
			v, err := item.GetResult()
			assert.NoError(t, err)
			u := v.(*User)
			items1 <- u
		}
		}
	}()

	opts = ravendb.NewSubscriptionWorkerOptions(id2)
	subscription2, err = store.Subscriptions().GetSubscriptionWorker(clazz, opts, "")
	assert.NoError(t, err)
	items2 := make(chan *User, 10)

	results2, err := subscription2.Run()
	assert.NoError(t, err)
	go func() {
		for items := range results2 {
			for _, item := range items {
				v, err := item.GetResult()
				assert.NoError(t, err)
				u := v.(*User)
				items2 <- u
			}
		}
	}()

	u, ok := getNextUser(items1, 0)
	assert.True(t, ok)
	assert.Equal(t, u.ID, "users/1")

	u, ok = getNextUser(items1, 0)
	assert.True(t, ok)
	assert.Equal(t, u.ID, "users/2")

	u, ok = getNextUser(items2, 0)
	assert.True(t, ok)
	assert.Equal(t, u.ID, "users/1")

	u, ok = getNextUser(items2, 0)
	assert.True(t, ok)
	assert.Equal(t, u.ID, "users/2")

	_ = subscription1.Close()
	subscription1 = nil

	{
		session, err := store.OpenSession("")
		assert.NoError(t, err)
		err = session.StoreWithID(&User{}, "users/3")
		assert.NoError(t, err)
		err = session.StoreWithID(&User{}, "users/4")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	u, ok = getNextUser(items2, 0)
	assert.True(t, ok)
	assert.Equal(t, u.ID, "users/3")

	u, ok = getNextUser(items2, 0)
	assert.True(t, ok)
	assert.Equal(t, u.ID, "users/4")
}

func TestSubscriptionsBasic(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	subscriptionsBasic_canReleaseSubscription(t, driver)
	subscriptionsBasic_shouldRespectMaxDocCountInBatch(t, driver)
	subscriptionsBasic_shouldStreamAllDocumentsAfterSubscriptionCreation(t, driver)
	subscriptionsBasic_shouldRespectCollectionCriteria(t, driver)
	subscriptionsBasic_willAcknowledgeEmptyBatches(t, driver)
	subscriptionsBasic_shouldStopPullingDocsAndCloseSubscriptionOnSubscriberErrorByDefault(t, driver)
	subscriptionsBasic_disposingOneSubscriptionShouldNotAffectOnNotificationsOfOthers(t, driver)
	subscriptionsBasic_shouldPullDocumentsAfterBulkInsert(t, driver)
	//subscriptionsBasic_canSetToIgnoreSubscriberErrors(t, driver)
	subscriptionsBasic_ravenDB_3452_ShouldStopPullingDocsIfReleased(t, driver)
	subscriptionsBasic_canDeleteSubscription(t, driver)

subscriptionsBasic_shouldThrowOnAttemptToOpenAlreadyOpenedSubscription(t, driver)

	subscriptionsBasic_shouldThrowWhenOpeningNoExistingSubscription(t, driver)
	subscriptionsBasic_shouldSendAllNewAndModifiedDocs(t, driver)
	subscriptionsBasic_ravenDB_3453_ShouldDeserializeTheWholeDocumentsAfterTypedSubscription(t, driver)
}
