package main

// Note: this file is just to make sure that the code for the book examples
// compile. This code is not supposed to be run

import (
	"fmt"
	"reflect"

	"github.com/ravendb/ravendb-go-client"
	"github.com/ravendb/ravendb-go-client/examples/northwind"
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

func createDocument(companyName, companyPhone, contactName, contactTitle string) error {
	newCompany := &northwind.Company{
		Name:  companyName,
		Phone: companyPhone,
		Contact: &northwind.Contact{
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

	var company *northwind.Company
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

func createRelatedDocuments(productName, supplierName, supplierPhone string) error {
	supplier := &northwind.Supplier{
		Name:  supplierName,
		Phone: supplierPhone,
	}

	category := &northwind.Category{
		Name:        "NoSQL Databases",
		Description: "Non-relational databases",
	}

	product := &northwind.Product{
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

	var product *northwind.Product
	err = session.Include("supplier").Load(&product, "products/34-A")
	if err != nil {
		return err
	}
	if product == nil {
		// not found
		return nil
	}

	var supplier *northwind.Supplier
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

func queryRelatedDocuments() error {
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	tp := reflect.TypeOf(&northwind.Order{})
	q := session.QueryCollectionForType(tp)
	// TODO: will this work?
	q = q.Include("Lines.Product")
	q = q.WhereNotEquals("ShippedAt", nil)
	var shippedOrders []*northwind.Order
	err = q.GetResults(&shippedOrders)
	if err != nil {
		return err
	}
	for _, shippedOrder := range shippedOrders {
		for i, line := range shippedOrder.Lines {
			productID := line.Product
			var product *northwind.Product
			err = session.Load(&product, productID)
			if err != nil {
				return err
			}
			product.UnitsOnOrder += shippedOrder.Lines[i].Quantity
		}
	}

	err = session.SaveChanges()
	if err != nil {
		return err
	}

	return nil
}

func indexRelatedDocuments(categoryName string) error {
	index := ravendb.NewIndexCreationTask("Products/ByCategoryName")
	// TODO: verify this is correct
	index.Map = `docs.Products.Select(product => new {
		CategoryName = (this.LoadDocument(product.Category, "Categories")).Name
	})
`
	session, err := globalDocumentStore.OpenSession("")
	if err != nil {
		return err
	}
	defer session.Close()

	var productsWithCategoryName []*northwind.Product
	q := session.QueryIndex(index.IndexName)
	q = q.WhereEquals("CategoryName", categoryName)
	err = q.GetResults(&productsWithCategoryName)
	if err != nil {
		return err
	}
	return nil
}

func main() {

}
