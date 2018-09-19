package tests

import (
	"testing"
)

func changesTest_singleDocumentChanges(t *testing.T) {}

func changesTest_changesWithHttps(t *testing.T) {}

func changesTest_allDocumentsChanges(t *testing.T) {
	/*
		var err error
		store := getDocumentStoreMust(t)
		defer store.Close()

		{
			// BlockingQueue<DocumentChange> changesList = new BlockingArrayQueue<>();

			changes := store.Changes()
			err = changes.EnsureConnectedNow()
			assert.NoError(t, err)

			// IChangesObservable<DocumentChange>
			observable, err := changes.ForAllDocuments()
			assert.NoError(t, err)

			{
				subscription := observable.Subscribe()
			}
		}
	*/

	/*
	   try (CleanCloseable subscription = observable.subscribe(Observers.create(changesList::add))) {

	       try (IDocumentSession session = store.openSession()) {
	           User user = new User();
	           session.store(user, "users/1");
	           session.saveChanges();
	       }

	       DocumentChange documentChange = changesList.poll(2, TimeUnit.SECONDS);
	       assertThat(documentChange)
	               .isNotNull();

	       assertThat(documentChange.getId())
	               .isEqualTo("users/1");

	       assertThat(documentChange.getType())
	               .isEqualTo(DocumentChangeTypes.PUT);

	       DocumentChange secondPoll = changesList.poll(1, TimeUnit.SECONDS);
	       assertThat(secondPoll)
	               .isNull();
	   }


	   // at this point we should be unsubscribed from changes on 'users/1'

	   try (IDocumentSession session = store.openSession()) {
	       User user = new User();
	       user.setName("another name");
	       session.store(user, "users/1");
	       session.saveChanges();
	   }

	   // it should be empty as we destroyed subscription
	   DocumentChange thirdPoll = changesList.poll(1, TimeUnit.SECONDS);
	   assertThat(thirdPoll)
	           .isNull();
	*/
}

func changesTest_singleIndexChanges(t *testing.T) {}

func changesTest_allIndexChanges(t *testing.T) {}

func changesTest_notificationOnWrongDatabase_ShouldNotCrashServer(t *testing.T) {}

func changesTest_resourcesCleanup(t *testing.T) {}

/*
   public static class UsersByName extends AbstractIndexCreationTask {
       public UsersByName() {

           map = "from c in docs.Users select new " +
                   " {" +
                   "    c.name, " +
                   "    count = 1" +
                   "}";

           reduce = "from result in results " +
                   "group result by result.name " +
                   "into g " +
                   "select new " +
                   "{ " +
                   "  name = g.Key, " +
                   "  count = g.Sum(x => x.count) " +
                   "}";
       }
   }
*/

func TestChanges(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// follows execution order of java tests
	changesTest_allDocumentsChanges(t)
	changesTest_singleDocumentChanges(t)
	changesTest_resourcesCleanup(t)
	changesTest_changesWithHttps(t)
	changesTest_singleIndexChanges(t)
	changesTest_notificationOnWrongDatabase_ShouldNotCrashServer(t)
	changesTest_allIndexChanges(t)
}
