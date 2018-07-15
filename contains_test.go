package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
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

			err = session.StoreEntity(user)
			assert.NoError(t, err)
		}

		userCreator("John", []string{"java", "c#"})
		userCreator("Tarzan", []string{"java", "go"})
		userCreator("Jane", []string{"pascal"})

		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	/*
			{
				session := openSessionMust(t, store)
				List<String> pascalOrGoDeveloperNames = session
				.query(UserWithFavs.class)
				.containsAny("favourites", Arrays.asList("pascal", "go"))
				.selectFields(String.class, "name")
				.toList();

		assertThat(pascalOrGoDeveloperNames)
				.hasSize(2)
				.contains("Jane")
				.contains("Tarzan");
			}*/

}

/*
   try (IDocumentSession session = store.openSession()) {
       List<String> pascalOrGoDeveloperNames = session
               .query(UserWithFavs.class)
               .containsAny("favourites", Arrays.asList("pascal", "go"))
               .selectFields(String.class, "name")
               .toList();

       assertThat(pascalOrGoDeveloperNames)
               .hasSize(2)
               .contains("Jane")
               .contains("Tarzan");
   }

   try (IDocumentSession session = store.openSession()) {
       List<String> javaDevelopers = session
               .query(UserWithFavs.class)
               .containsAll("favourites", Collections.singletonList("java"))
               .selectFields(String.class, "name")
               .toList();

       assertThat(javaDevelopers)
               .hasSize(2)
               .contains("John")
               .contains("Tarzan");
   }
*/

func TestContains(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_contains_go.txt")
	}

	createTestDriver()
	defer deleteTestDriver()

	// TODO: re-enable when finished
	//containsTestcontainsTest(t)
}
