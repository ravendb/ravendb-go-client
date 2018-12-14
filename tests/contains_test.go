package tests

import (
	"reflect"
	"testing"

	"github.com/ravendb/ravendb-go-client"
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
	store := getDocumentStoreMust(t, driver)
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

		q := session.QueryOld(reflect.TypeOf(&UserWithFavs{}))
		q = q.ContainsAny("Favourites", []interface{}{"pascal", "go"})
		q = q.SelectFields(reflect.TypeOf(""), "Name")
		pascalOrGoDeveloperNames, err := q.ToListOld()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(pascalOrGoDeveloperNames))
		assert.True(t, ravendb.InterfaceArrayContains(pascalOrGoDeveloperNames, "Jane"))
		assert.True(t, ravendb.InterfaceArrayContains(pascalOrGoDeveloperNames, "Tarzan"))

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q := session.QueryOld(reflect.TypeOf(&UserWithFavs{}))
		q = q.ContainsAll("Favourites", []interface{}{"java"})
		q = q.SelectFields(reflect.TypeOf(""), "Name")
		javaDevelopers, err := q.ToListOld()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(javaDevelopers))
		assert.True(t, ravendb.InterfaceArrayContains(javaDevelopers, "John"))
		assert.True(t, ravendb.InterfaceArrayContains(javaDevelopers, "Tarzan"))

		session.Close()
	}

}

func TestContains(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	containsTestcontainsTest(t, driver)
}
