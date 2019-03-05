package main

// Note: this file is just to make sure that the code for the book examples
// compile. This code is not supposed to be run

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

type Supplier struct {
	ID    string
	Name  string `json:"name,omitempty"`
	Phone string `json:"phone,omitempty"`
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

func editDocumentChapter(companyName string) error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	var company *Company
	err = session.Load(&company, "companies/5-A")
	if err != nil {
		return err
	}
	// if not found, company is not changed and will remain nil
	if company == nil {
		return nil
	}
	company.Name = companyName

	err = session.SaveChanges()
	if err != nil {
		return err
	}

	return nil
}

func deleteDocumentChapter(documentId string) error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	err = session.DeleteByID(documentId, nil)
	if err != nil {
		return err
	}

	err = session.SaveChanges()
	if err != nil {
		return err
	}

	return nil
}

type Category struct {
	ID          string
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Product struct {
	ID           string
	Name         string  `json:"name"`
	Supplier     string  `json:"supplier"`
	Category     string  `json:"category"`
	PricePerUnit float64 `json:"pricePerUnit"`
}

func createRelatedDocuments(productName, supplierName, supplierPhone string) error {
	supplier := &Supplier{
		Name:  supplierName,
		Phone: supplierPhone,
	}

	category := &Category{
		Name:        "NoSQL Databases",
		Description: "Non-relational databases",
	}

	product := &Product{
		Name: productName,
	}

	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	err = session.Store(supplier)
	if err != nil {
		return err
	}
	err = session.Store(category)
	if err != nil {
		return err
	}

	product.Supplier = supplier.ID
	product.Category = category.ID

	err = session.Store(product)
	if err != nil {
		return err
	}

	err = session.SaveChanges()
	if err != nil {
		return err
	}

	return nil
}

func loadRelatedDocuments(pricePerUnit float64, phone string) error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	var product *Product
	err = session.Include("supplier").Load(&product, "products/34-A")
	if err != nil {
		return err
	}
	if product == nil {
		// not found
		return nil
	}

	var supplier *Supplier
	err = session.Load(&supplier, product.Supplier)
	if err != nil || supplier == nil {
		return err
	}

	product.PricePerUnit = pricePerUnit
	supplier.Phone = phone

	err = session.SaveChanges()
	if err != nil {
		return err
	}

	return nil
}

func main() {

}
