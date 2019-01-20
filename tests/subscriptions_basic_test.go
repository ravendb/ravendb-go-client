package tests

import (
	"reflect"
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

const (
	_reasonableWaitTime = time.Second * 5 // TODO: is it 60 seconds in Java?
)

func subscriptionsBasic_canDeleteSubscription(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	id1, err := store.Subscriptions.CreateForType(reflect.TypeOf(&User{}), nil, "")
	assert.NoError(t, err)
	id2, err := store.Subscriptions.CreateForType(reflect.TypeOf(&User{}), nil, "")
	assert.NoError(t, err)

	subscriptions, err := store.Subscriptions.GetSubscriptions(0, 5, "")
	assert.NoError(t, err)
	assert.Equal(t, len(subscriptions), 2)
	err = store.Subscriptions.Delete(id1, "")
	assert.NoError(t, err)
	err = store.Subscriptions.Delete(id2, "")
	assert.NoError(t, err)

	subscriptions, err = store.Subscriptions.GetSubscriptions(0, 5, "")
	assert.NoError(t, err)
	assert.Equal(t, len(subscriptions), 0)
}

func subscriptionsBasic_shouldThrowWhenOpeningNoExisingSubscription(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	clazz := reflect.TypeOf(&map[string]interface{}{})
	opts, err := ravendb.NewSubscriptionWorkerOptions("1")
	assert.NoError(t, err)
	subscription, err := store.Subscriptions.GetSubscriptionWorker(clazz, opts, "")
	assert.NoError(t, err)
	fn := func(x *ravendb.SubscriptionBatch) error {
		// no-op
		return nil
	}

	res, err := subscription.Run(fn)
	assert.NoError(t, err)
	_, err = res.Get()
	assert.NotNil(t, err)
	_, ok := err.(*ravendb.SubscriptionDoesNotExistError)
	assert.True(t, ok)
}

func subscriptionsBasic_shouldThrowOnAttemptToOpenAlreadyOpenedSubscription(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	id, err := store.Subscriptions.CreateForType(reflect.TypeOf(&User{}), nil, "")
	assert.NoError(t, err)

	{
		clazz := reflect.TypeOf(map[string]interface{}{})
		opts, err := ravendb.NewSubscriptionWorkerOptions(id)
		assert.NoError(t, err)
		subscription, err := store.Subscriptions.GetSubscriptionWorker(clazz, opts, "")
		assert.NoError(t, err)

		{
			session, err := store.OpenSession()
			assert.NoError(t, err)
			err = session.Store(&User{})
			assert.NoError(t, err)
			err = session.SaveChanges()
			assert.NoError(t, err)

			session.Close()
		}

		semaphore := make(chan bool)
		fn := func(x *ravendb.SubscriptionBatch) error {
			semaphore <- true
			return nil
		}
		_, err = subscription.Run(fn)
		assert.NoError(t, err)

		select {
		case <-semaphore:
			// no-op
		case <-time.After(_reasonableWaitTime):
			// no-op
		}

		options2, err := ravendb.NewSubscriptionWorkerOptions(id)
		assert.NoError(t, err)
		options2.Strategy = ravendb.SubscriptionOpeningStrategyOpenIfFree

		{
			secondSubscription, err := store.Subscriptions.GetSubscriptionWorker(clazz, options2, "")
			assert.NoError(t, err)
			fn := func(x *ravendb.SubscriptionBatch) error {
				// no-op
				return nil
			}
			future, err := secondSubscription.Run(fn)
			assert.NoError(t, err)
			_, err = future.Get()
			_, ok := err.(*ravendb.SubscriptionInUseError)
			assert.True(t, ok)
		}
	}

}

func subscriptionsBasic_shouldStreamAllDocumentsAfterSubscriptionCreation(t *testing.T, driver *RavenTestDriver) {
}

func subscriptionsBasic_shouldSendAllNewAndModifiedDocs(t *testing.T, driver *RavenTestDriver) {
}

func subscriptionsBasic_shouldRespectMaxDocCountInBatch(t *testing.T, driver *RavenTestDriver) {
}

func subscriptionsBasic_shouldRespectCollectionCriteria(t *testing.T, driver *RavenTestDriver) {
}

func subscriptionsBasic_willAcknowledgeEmptyBatches(t *testing.T, driver *RavenTestDriver) {
}

func subscriptionsBasic_canReleaseSubscription(t *testing.T, driver *RavenTestDriver) {
}

func subscriptionsBasic_shouldPullDocumentsAfterBulkInsert(t *testing.T, driver *RavenTestDriver) {
}

func subscriptionsBasic_shouldStopPullingDocsAndCloseSubscriptionOnSubscriberErrorByDefault(t *testing.T, driver *RavenTestDriver) {
}

func subscriptionsBasic_canSetToIgnoreSubscriberErrors(t *testing.T, driver *RavenTestDriver) {
}

func subscriptionsBasic_ravenDB_3452_ShouldStopPullingDocsIfReleased(t *testing.T, driver *RavenTestDriver) {
}

func subscriptionsBasic_ravenDB_3453_ShouldDeserializeTheWholeDocumentsAfterTypedSubscription(t *testing.T, driver *RavenTestDriver) {
}

func subscriptionsBasic_disposingOneSubscriptionShouldNotAffectOnNotificationsOfOthers(t *testing.T, driver *RavenTestDriver) {
}

func TestSubscriptionsBasic(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests

	// TODO: arrange in Java order
	subscriptionsBasic_canDeleteSubscription(t, driver)
	subscriptionsBasic_shouldThrowWhenOpeningNoExisingSubscription(t, driver)
	subscriptionsBasic_shouldThrowOnAttemptToOpenAlreadyOpenedSubscription(t, driver)

	/*
		subscriptionsBasic_shouldStreamAllDocumentsAfterSubscriptionCreation(t, driver)
		subscriptionsBasic_shouldSendAllNewAndModifiedDocs(t, driver)
		subscriptionsBasic_shouldRespectMaxDocCountInBatch(t, driver)
		subscriptionsBasic_shouldRespectCollectionCriteria(t, driver)
		subscriptionsBasic_willAcknowledgeEmptyBatches(t, driver)
		subscriptionsBasic_canReleaseSubscription(t, driver)
		subscriptionsBasic_shouldPullDocumentsAfterBulkInsert(t, driver)
		subscriptionsBasic_shouldStopPullingDocsAndCloseSubscriptionOnSubscriberErrorByDefault(t, driver)
		subscriptionsBasic_canSetToIgnoreSubscriberErrors(t, driver)
		subscriptionsBasic_ravenDB_3452_ShouldStopPullingDocsIfReleased(t, driver)
		subscriptionsBasic_ravenDB_3453_ShouldDeserializeTheWholeDocumentsAfterTypedSubscription(t, driver)
		subscriptionsBasic_disposingOneSubscriptionShouldNotAffectOnNotificationsOfOthers(t, driver)
	*/
}
