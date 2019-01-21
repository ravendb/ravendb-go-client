package tests

import (
	"fmt"
	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strconv"
	"testing"
)

func revisionsSubscriptions_plainRevisionsSubscriptions(t *testing.T, driver *RavenTestDriver) {
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
					u := result.Current.(*User)
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
	/*
	   try (IDocumentStore store = getDocumentStore()) {
	       String subscriptionId = store.subscriptions().createForRevisions(User.class);

	       RevisionsCollectionConfiguration defaultCollection = new RevisionsCollectionConfiguration();
	       defaultCollection.setDisabled(false);
	       defaultCollection.setMinimumRevisionsToKeep(5l);

	       RevisionsCollectionConfiguration usersConfig = new RevisionsCollectionConfiguration();
	       usersConfig.setDisabled(false);

	       RevisionsCollectionConfiguration donsConfig = new RevisionsCollectionConfiguration();
	       donsConfig.setDisabled(false);

	       RevisionsConfiguration configuration = new RevisionsConfiguration();
	       configuration.setDefaultConfig(defaultCollection);

	       HashMap<String, RevisionsCollectionConfiguration> perCollectionConfig = new HashMap<>();
	       perCollectionConfig.put("Users", usersConfig);
	       perCollectionConfig.put("Dons", donsConfig);

	       configuration.setCollections(perCollectionConfig);

	       ConfigureRevisionsOperation operation = new ConfigureRevisionsOperation(configuration);

	       store.maintenance().send(operation);


	       for (int j = 0; j < 10; j++) {
	           try (IDocumentSession session = store.openSession()) {
	               User user = new User();
	               user.setName("users1 ver " + j);
	               user.setAge(j);
	               session.store(user, "users/1");

	               Company company = new Company();
	               company.setName("dons1 ver " + j);
	               session.store(company, "dons/1");

	               session.saveChanges();
	           }
	       }

	       try (SubscriptionWorker<Revision<User>> sub = store.subscriptions().getSubscriptionWorkerForRevisions(User.class, new SubscriptionWorkerOptions(subscriptionId))) {
	           Semaphore mre = new Semaphore(0);
	           Set<String> names = new HashSet<>();

	           final AtomicInteger maxAge = new AtomicInteger(-1);

	           sub.run(a -> {
	               for (SubscriptionBatch.Item<Revision<User>> item : a.getItems()) {
	                   Revision<User> x = item.getResult();
	                   if (x.getCurrent().getAge() > maxAge.get() && x.getCurrent().getAge() > Optional.ofNullable(x.getPrevious()).map(y -> y.getAge()).orElse(-1)) {
	                       names.add(Optional.ofNullable(x.getCurrent()).map(y -> y.getName()).orElse(null) + " "
	                               +  Optional.ofNullable(x.getPrevious()).map(y -> y.getName()).orElse(null));
	                       maxAge.set(x.getCurrent().getAge());
	                   }

	                   if (names.size() == 10) {
	                       mre.release();
	                   }
	               }
	           });

	           assertThat(mre.tryAcquire(_reasonableWaitTime, TimeUnit.SECONDS))
	                   .isTrue();

	       }
	   }
	*/
}

func TestRevisionsSubscriptions(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	revisionsSubscriptions_plainRevisionsSubscriptions(t, driver)
	revisionsSubscriptions_plainRevisionsSubscriptionsCompareDocs(t, driver)

	// matches order of Java tests
}
