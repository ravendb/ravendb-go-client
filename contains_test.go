package ravendb

import (
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

func (u *UserWithFavs) getId() string {
	return u.ID
}

func (u *UserWithFavs) setId(id string) {
	u.ID = id
}

func (u *UserWithFavs) getName() string {
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

		q := session.query(getTypeOf(&UserWithFavs{}))
		q = q.containsAny("Favourites", []Object{"pascal", "go"})
		q = q.selectFields(getTypeOf(""), "Name")
		pascalOrGoDeveloperNames, err := q.toList()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(pascalOrGoDeveloperNames))
		assert.True(t, interfaceArrayContains(pascalOrGoDeveloperNames, "Jane"))
		assert.True(t, interfaceArrayContains(pascalOrGoDeveloperNames, "Tarzan"))

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q := session.query(getTypeOf(&UserWithFavs{}))
		q = q.containsAll("Favourites", []Object{"java"})
		q = q.selectFields(getTypeOf(""), "Name")
		javaDevelopers, err := q.toList()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(javaDevelopers))
		assert.True(t, interfaceArrayContains(javaDevelopers, "John"))
		assert.True(t, interfaceArrayContains(javaDevelopers, "Tarzan"))

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
