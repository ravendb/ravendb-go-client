package main

import (
	"fmt"

	"github.com/ravendb/ravendb-go-client"
)

var (
	serverURL    = "http://localhost:8080"
	databaseName = "YourDatabaseName"
)

var (
	globalDocumentStore *ravendb.DocumentStore
)

func createDocumentStore() (*ravendb.DocumentStore, error) {
	if globalDocumentStore != nil {
		return globalDocumentStore, nil
	}
	urls := []string{serverURL}
	store := ravendb.NewDocumentStore(urls, databaseName)
	err := store.Initialize()
	if err != nil {
		return nil, err
	}
	globalDocumentStore = store
	return globalDocumentStore, nil
}

// Contact describes a contact
type Contact struct {
	Name  string `json:"name,omitempty"`
	Title string `json:"title,omitempty"`
}

// Company describes a company
type Company struct {
	ID      string
	Name    string   `json:"name,omitempty"`
	Phone   string   `json:"phone,omitempty"`
	Contact *Contact `json:"contact"`
}

func createDocument(companyName, companyPhone, contactName, contactTitle string) error {
	newCompany := &Company{
		Name:  companyName,
		Phone: companyPhone,
		Contact: &Contact{
			Name:  contactName,
			Title: contactTitle,
		},
	}

	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	err = session.Store(newCompany)
	if err != nil {
		return err
	}

	theNewDocumentID := newCompany.ID

	err = session.SaveChanges()
	if err != nil {
		return err
	}

	fmt.Printf("Document id: %s\n", theNewDocumentID)
	return nil
}

func sessionChapter() error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	//   Run your business logic:
	//
	//   Store documents
	//   Load and Modify documents
	//   Query indexes & collections
	//   Delete documents
	//   .... etc.

	err = session.SaveChanges()
	if err != nil {
		return err
	}

	// make sure to close the sessoin
	session.Close()

	return nil
}
func main() {

}
