package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func documentStreaming_canStreamDocumentsStartingWith(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		for i := 0; i < 200; i++ {
			user := &User{}
			err = session.Store(user)
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	/*
	       int count = 0;

	       try (IDocumentSession session = store.openSession()) {
	           try (CloseableIterator<StreamResult<User>> reader = session.advanced().stream(User.class, "users/")) {
	               while (reader.hasNext()) {
	                   count++;
	                   User user = reader.next().getDocument();
	                   assertThat(user)
	                           .isNotNull();
	               }
	           }
	       }

	       assertThat(count)
	               .isEqualTo(200);
	   }
	*/
}

func documentStreaming_streamWithoutIterationDoesntLeakConnection(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		for i := 0; i < 200; i++ {
			user := &User{}
			err = session.Store(user)
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	/*
	   for (int i = 0; i < 5; i++) {
	       try (IDocumentSession session = store.openSession()) {

	           try (CloseableIterator<StreamResult<User>> reader = session.advanced().stream(User.class, "users/")) {
	               // don't iterate
	           }
	       }
	   }
	*/
}

func TestDocumentStreaming(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	documentStreaming_canStreamDocumentsStartingWith(t)
	documentStreaming_streamWithoutIterationDoesntLeakConnection(t)
}
