package tests

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type UserWithFavs struct {
	ID         string
	Name       string
	Favourites []string
}

func NewUserWithFavs() *UserWithFavs {
	return &UserWithFavs{}
}

func containsTestcontainsTest(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		userCreator := func(name string, favs []string) {
			user := NewUserWithFavs()
			user.Name = name
			user.Favourites = favs

			err = session.Store(user)
			assert.NoError(t, err)
		}

		userCreator("John", []string{"java", "c#"})
		userCreator("Tarzan", []string{"java", "go"})
		userCreator("Jane", []string{"pascal"})

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q := session.QueryType(reflect.TypeOf(&UserWithFavs{}))
		q = q.ContainsAny("Favourites", []interface{}{"pascal", "go"})
		q = q.SelectFields("Name")
		var pascalOrGoDeveloperNames []string
		err = q.GetResults(&pascalOrGoDeveloperNames)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(pascalOrGoDeveloperNames))
		assert.Equal(t, pascalOrGoDeveloperNames[0], "Tarzan")
		assert.Equal(t, pascalOrGoDeveloperNames[1], "Jane")

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q := session.QueryType(reflect.TypeOf(&UserWithFavs{}))
		q = q.ContainsAll("Favourites", []interface{}{"java"})
		q = q.SelectFields("Name")
		var javaDevelopers []string
		err = q.GetResults(&javaDevelopers)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(javaDevelopers))
		assert.Equal(t, javaDevelopers[0], "John")
		assert.Equal(t, javaDevelopers[1], "Tarzan")

		session.Close()
	}

}

func TestContains(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	containsTestcontainsTest(t, driver)
}
