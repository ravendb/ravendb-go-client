package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

type UserWithFavs struct {
	ID         string
	Name       string
	Favourites []string
}

func NewUserWithFavs() *UserWithFavs {
	return &UserWithFavs{}
}

func (u *UserWithFavs) getId() string {
	return u.ID
}

func (u *UserWithFavs) setId(id string) {
	u.ID = id
}

func (u *UserWithFavs) GetName() string {
	return u.Name
}

func (u *UserWithFavs) setName(name string) {
	u.Name = name
}

func (u *UserWithFavs) getFavourites() []string {
	return u.Favourites
}

func (u *UserWithFavs) setFavourites(favourites []string) {
	u.Favourites = favourites
}

func containsTestcontainsTest(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		userCreator := func(name string, favs []string) {
			user := NewUserWithFavs()
			user.setName(name)
			user.setFavourites(favs)

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

		q := session.Query(ravendb.GetTypeOf(&UserWithFavs{}))
		q = q.ContainsAny("Favourites", []ravendb.Object{"pascal", "go"})
		q = q.SelectFields(ravendb.GetTypeOf(""), "Name")
		pascalOrGoDeveloperNames, err := q.ToList()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(pascalOrGoDeveloperNames))
		assert.True(t, ravendb.InterfaceArrayContains(pascalOrGoDeveloperNames, "Jane"))
		assert.True(t, ravendb.InterfaceArrayContains(pascalOrGoDeveloperNames, "Tarzan"))

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q := session.Query(ravendb.GetTypeOf(&UserWithFavs{}))
		q = q.ContainsAll("Favourites", []ravendb.Object{"java"})
		q = q.SelectFields(ravendb.GetTypeOf(""), "Name")
		javaDevelopers, err := q.ToList()
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

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	containsTestcontainsTest(t)
}
