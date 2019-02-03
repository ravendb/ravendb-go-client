package ravendb

import (
	"io"
	"reflect"
	"sync"
)

// DocumentSubscriptions describes document subscriptions
type DocumentSubscriptions struct {
	store         *DocumentStore
	subscriptions map[io.Closer]bool
	mu            sync.Mutex // protects subscriptions
}

// NewDocumentSubscriptions returns new DocumentSubscriptions
func NewDocumentSubscriptions(store *DocumentStore) *DocumentSubscriptions {
	return &DocumentSubscriptions{
		store:         store,
		subscriptions: map[io.Closer]bool{},
	}
}

// Create creates a data subscription in a database. The subscription will expose all documents that match the specified subscription options for a given type.
func (s *DocumentSubscriptions) Create(options *SubscriptionCreationOptions, database string) (string, error) {
	if options == nil {
		return "", newIllegalArgumentError("Cannot create a subscription if options is nil")
	}

	if options.Query == "" {
		return "", newIllegalArgumentError("Cannot create a subscription if the script is empty string")
	}

	if database == "" {
		database = s.store.GetDatabase()
	}
	requestExecutor := s.store.GetRequestExecutor(database)

	command := newCreateSubscriptionCommand(s.store.GetConventions(), options, "")
	if err := requestExecutor.ExecuteCommand(command, nil); err != nil {
		return "", err
	}

	return command.Result.Name, nil
}

// CreateForType creates a data subscription in a database. The subscription will expose all documents that match the specified subscription options for a given type.
func (s *DocumentSubscriptions) CreateForType(clazz reflect.Type, options *SubscriptionCreationOptions, database string) (string, error) {
	if options == nil {
		options = &SubscriptionCreationOptions{}

	}
	creationOptions := &SubscriptionCreationOptions{
		Name:         options.Name,
		ChangeVector: options.ChangeVector,
	}

	opts := s.ensureCriteria(creationOptions, clazz, false)
	return s.Create(opts, database)
}

// CreateForRevisions creates a data subscription in a database. The subscription will expose all documents that match the specified subscription options for a given type.
func (s *DocumentSubscriptions) CreateForRevisions(clazz reflect.Type, options *SubscriptionCreationOptions, database string) (string, error) {
	if options == nil {
		options = &SubscriptionCreationOptions{}
	}
	creationOptions := &SubscriptionCreationOptions{
		Name:         options.Name,
		ChangeVector: options.ChangeVector,
	}

	opts := s.ensureCriteria(creationOptions, clazz, true)
	return s.Create(opts, database)
}

func (s *DocumentSubscriptions) ensureCriteria(criteria *SubscriptionCreationOptions, clazz reflect.Type, revisions bool) *SubscriptionCreationOptions {
	if criteria == nil {
		criteria = &SubscriptionCreationOptions{}
	}

	collectionName := s.store.GetConventions().GetCollectionName(clazz)
	if criteria.Query == "" {
		if revisions {
			criteria.Query = "from " + collectionName + " (Revisions = true) as doc"
		} else {
			criteria.Query = "from " + collectionName + " as doc"
		}
	}

	return criteria
}

// GetSubscriptionWorker opens a subscription and starts pulling documents since a last processed document for that subscription.
// The connection options determine client and server cooperation rules like document batch sizes or a timeout in a matter of which a client
// needs to acknowledge that batch has been processed. The acknowledgment is sent after all documents are processed by subscription's handlers.
func (s *DocumentSubscriptions) GetSubscriptionWorker(clazz reflect.Type, options *SubscriptionWorkerOptions, database string) (*SubscriptionWorker, error) {
	if err := s.store.assertInitialized(); err != nil {
		return nil, err
	}

	if options == nil {
		return nil, newIllegalStateError("Cannot open a subscription if options are null")
	}

	subscription, err := NewSubscriptionWorker(clazz, options, false, s.store, database)
	if err != nil {
		return nil, err
	}
	fn := func(worker *SubscriptionWorker) {
		s.mu.Lock()
		delete(s.subscriptions, worker)
		s.mu.Unlock()
	}
	subscription.onClosed = fn
	s.mu.Lock()
	s.subscriptions[subscription] = true
	s.mu.Unlock()

	return subscription, nil
}

// GetSubscriptionWorkerForRevisions opens a subscription and starts pulling documents since a last processed document for that subscription.
// The connection options determine client and server cooperation rules like document batch sizes or a timeout in a matter of which a client
// needs to acknowledge that batch has been processed. The acknowledgment is sent after all documents are processed by subscription's handlers.
func (s *DocumentSubscriptions) GetSubscriptionWorkerForRevisions(clazz reflect.Type, options *SubscriptionWorkerOptions, database string) (*SubscriptionWorker, error) {
	subscription, err := NewSubscriptionWorker(clazz, options, true, s.store, database)
	if err != nil {
		return nil, err
	}

	fn := func(sender *SubscriptionWorker) {
		s.mu.Lock()
		delete(s.subscriptions, sender)
		s.mu.Unlock()
	}

	subscription.onClosed = fn
	s.mu.Lock()
	s.subscriptions[subscription] = true
	s.mu.Unlock()

	return subscription, nil
}

// GetSubscriptions downloads a list of all existing subscriptions in a database.
func (s *DocumentSubscriptions) GetSubscriptions(start int, take int, database string) ([]*SubscriptionState, error) {
	if database == "" {
		database = s.store.GetDatabase()
	}
	requestExecutor := s.store.GetRequestExecutor(database)

	command := newGetSubscriptionsCommand(start, take)
	if err := requestExecutor.ExecuteCommand(command, nil); err != nil {
		return nil, err
	}

	return command.Result, nil
}

// Delete deletes a subscription.
func (s *DocumentSubscriptions) Delete(name string, database string) error {
	if database == "" {
		database = s.store.GetDatabase()
	}
	requestExecutor := s.store.GetRequestExecutor(database)

	command := newDeleteSubscriptionCommand(name)
	return requestExecutor.ExecuteCommand(command, nil)
}

// GetSubscriptionState returns subscription definition and it's current state
func (s *DocumentSubscriptions) GetSubscriptionState(subscriptionName string, database string) (*SubscriptionState, error) {
	if subscriptionName == "" {
		return nil, newIllegalArgumentError("SubscriptionName cannot be null")
	}

	if database == "" {
		database = s.store.GetDatabase()
	}
	requestExecutor := s.store.GetRequestExecutor(database)

	command := newGetSubscriptionStateCommand(subscriptionName)
	if err := requestExecutor.ExecuteCommand(command, nil); err != nil {
		return nil, err
	}
	return command.Result, nil
}

// Close closes subscriptions
func (s *DocumentSubscriptions) Close() error {
	if len(s.subscriptions) == 0 {
		return nil
	}

	var err error
	for subscription := range s.subscriptions {
		err2 := subscription.Close()
		if err2 != nil {
			err = err2
		}
	}
	return err
}

// DropConnection forces server to close current client subscription connection to the server
func (s *DocumentSubscriptions) DropConnection(name string, database string) error {
	if database == "" {
		database = s.store.GetDatabase()
	}
	requestExecutor := s.store.GetRequestExecutor(database)

	command := newDropSubscriptionConnectionCommand(name)
	return requestExecutor.ExecuteCommand(command, nil)
}
