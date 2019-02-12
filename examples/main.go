package main

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/ravendb/ravendb-go-client/examples/northwind"

	ravendb "github.com/ravendb/ravendb-go-client"
)

// "Demo" is a Northwind sample database
// You can browse its content via web interface at
// http://live-test.ravendb.net/studio/index.html#databases/documents?&database=Demo
var (
	dbName = "Demo"
)

func getDocumentStore(databaseName string) (*ravendb.DocumentStore, error) {
	serverNodes := []string{"http://live-test.ravendb.net"}
	store := ravendb.NewDocumentStore(serverNodes, databaseName)
	if err := store.Initialize(); err != nil {
		return nil, err
	}
	return store, nil
}

func openSession(databaseName string) (*ravendb.DocumentStore, *ravendb.DocumentSession, error) {
	store, err := getDocumentStore(dbName)
	if err != nil {
		return nil, nil, fmt.Errorf("getDocumentStore() failed with %s\n", err)
	}

	session, err := store.OpenSession("")
	if err != nil {
		return nil, nil, fmt.Errorf("store.OpenSession() failed with %s\n", err)
	}
	return store, session, nil
}

func loadUpdateSave() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	var e *northwind.Employee
	err = session.Load(&e, "employees/7-A")
	if err != nil {
		log.Fatalf("session.Load() failed with %s\n", err)
	}

	origName := e.FirstName
	e.FirstName = e.FirstName + "Changed"
	err = session.Store(e)
	if err != nil {
		log.Fatalf("session.Store() failed with %s\n", err)
	}

	err = session.SaveChanges()
	if err != nil {
		log.Fatalf("session.SaveChanges() failed with %s\n", err)
	}

	var e2 *northwind.Employee
	err = session.Load(&e2, "employees/7-A")
	if err != nil {
		log.Fatalf("session.Load() failed with %s\n", err)
	}
	fmt.Printf("Updated Employee.FirstName from '%s' to '%s'\n", origName, e2.FirstName)
}

func crudStore() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	product := &northwind.Product{
		Name:         "iPhone X",
		PricePerUnit: 999.99,
		Category:     "electronis",
		ReorderLevel: 15,
	}
	err = session.Store(product)
	if err != nil {
		log.Fatalf("session.Store() failed with %s\n", err)
	}
	fmt.Printf("Product ID: %s\n", product.ID)
	err = session.SaveChanges()
	if err != nil {
		log.Fatalf("session.SaveChanges() failed with %s\n", err)
	}
}

func crudLoad() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	var e *northwind.Employee
	err = session.Load(&e, "employees/7-A")
	if err != nil {
		log.Fatalf("session.Load() failed with %s\n", err)
	}
	spew.Fdump(os.Stdout, e)
	// fmt.Printf("employee: %+v\n", e)
}

func crudLoadWithIncludes() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// load employee with id "employees/7-A" and entity whose id is ReportsTo
	var e *northwind.Employee
	err = session.Include("ReportsTo").Load(&e, "employees/5-A")
	if err != nil {
		log.Fatalf("session.Load() failed with %s\n", err)
	}
	if e.ReportsTo == "" {
		fmt.Printf("Employee with id employees/5-A doesn't report to anyone\n")
		return
	}

	numRequests := session.GetNumberOfRequests()
	var reportsTo *northwind.Employee
	err = session.Load(&reportsTo, e.ReportsTo)
	if err != nil {
		log.Fatalf("session.Load() failed with %s\n", err)
	}
	if numRequests != session.GetNumberOfRequests() {
		fmt.Printf("Something's wrong, this shouldn't send a request to the server\n")
	} else {
		fmt.Printf("Loading e.ReportsTo employee didn't require a new request to the server because we've loaded it in original requests thanks to using Include functionality\n")
	}
}

func crudUpdate() {
	store, err := getDocumentStore(dbName)
	if err != nil {
		log.Fatalf("getDocumentStore() failed with %s\n", err)
	}
	defer store.Close()

	var productID string
	{
		session, err := store.OpenSession("")
		if err != nil {
			log.Fatalf("store.OpenSession() failed with %s\n", err)
		}

		product := &northwind.Product{
			Name:         "iPhone X",
			PricePerUnit: 999.99,
			Category:     "electronis",
			ReorderLevel: 15,
		}
		err = session.Store(product)
		if err != nil {
			log.Fatalf("session.Store() failed with %s\n", err)
		}
		productID = product.ID
		err = session.SaveChanges()
		if err != nil {
			log.Fatalf("session.SaveChanges() failed with %s\n", err)
		}

		session.Close()
	}

	var origPrice float64
	var newPrice float64
	{
		session, err := store.OpenSession("")
		if err != nil {
			log.Fatalf("store.OpenSession() failed with %s\n", err)
		}

		// load entity from the server
		var p *northwind.Product
		err = session.Load(&p, productID)
		if err != nil {
			log.Fatalf("session.Load() failed with %s\n", err)
		}

		// update price
		origPrice = p.PricePerUnit
		newPrice = origPrice + 10
		p.PricePerUnit = newPrice
		err = session.Store(p)
		if err != nil {
			log.Fatalf("session.Store() failed with %s\n", err)
		}

		// persist changes on the server
		err = session.SaveChanges()
		if err != nil {
			log.Fatalf("session.SaveChanges() failed with %s\n", err)
		}

		session.Close()
	}

	{
		session, err := store.OpenSession("")
		if err != nil {
			log.Fatalf("store.OpenSession() failed with %s\n", err)
		}
		var p *northwind.Product
		err = session.Load(&p, productID)
		if err != nil {
			log.Fatalf("session.Load() failed with %s\n", err)
		}

		if p.PricePerUnit != newPrice {
			fmt.Printf("Error: a change to PricePerUnit was not persisted (is %v should be %v)\n", newPrice, p.PricePerUnit)
		} else {
			fmt.Printf("Updated the price from %v to %v\n", origPrice, newPrice)
		}
		session.Close()
	}
}

func crudDeleteUsingID() {
	store, err := getDocumentStore(dbName)
	if err != nil {
		log.Fatalf("getDocumentStore() failed with %s\n", err)
	}
	defer store.Close()

	var productID string
	{
		session, err := store.OpenSession("")
		if err != nil {
			log.Fatalf("store.OpenSession() failed with %s\n", err)
		}

		product := &northwind.Product{
			Name:         "iPhone X",
			PricePerUnit: 999.99,
			Category:     "electronis",
		}
		err = session.Store(product)
		if err != nil {
			log.Fatalf("session.Store() failed with %s\n", err)
		}
		productID = product.ID
		err = session.SaveChanges()
		if err != nil {
			log.Fatalf("session.SaveChanges() failed with %s\n", err)
		}

		session.Close()
	}

	{
		session, err := store.OpenSession("")
		if err != nil {
			log.Fatalf("store.OpenSession() failed with %s\n", err)
		}

		err = session.Delete(productID)
		if err != nil {
			log.Fatalf("session.Delete() failed with %s\n", err)
		}

		err = session.SaveChanges()
		if err != nil {
			log.Fatalf("session.SaveChanges() failed with %s\n", err)
		}

		session.Close()
	}

	{
		session, err := store.OpenSession("")
		if err != nil {
			log.Fatalf("store.OpenSession() failed with %s\n", err)
		}

		// try to load deleted entity from the server
		var p *northwind.Product
		err = session.Load(&p, productID)
		if err != nil {
			log.Fatalf("session.Load() failed with %s\n", err)
		}

		if p == nil {
			fmt.Printf("Success: we deleted Product with this id so we get nil when we try to load it\n")
		} else {
			fmt.Printf("Error: this entity was deleted so we shouldn't be able to load it\n")
		}

		session.Close()
	}
}

func crudDeleteUsingEntity() {
	store, err := getDocumentStore(dbName)
	if err != nil {
		log.Fatalf("getDocumentStore() failed with %s\n", err)
	}
	defer store.Close()

	// store a new entity and remember its id
	var productID string
	{
		session, err := store.OpenSession("")
		if err != nil {
			log.Fatalf("store.OpenSession() failed with %s\n", err)
		}

		product := &northwind.Product{
			Name:         "iPhone X",
			PricePerUnit: 999.99,
			Category:     "electronis",
		}
		err = session.Store(product)
		if err != nil {
			log.Fatalf("session.Store() failed with %s\n", err)
		}
		productID = product.ID
		err = session.SaveChanges()
		if err != nil {
			log.Fatalf("session.SaveChanges() failed with %s\n", err)
		}

		session.Close()
	}

	// delete the entity
	{
		session, err := store.OpenSession("")
		if err != nil {
			log.Fatalf("store.OpenSession() failed with %s\n", err)
		}

		var p *northwind.Product
		err = session.Load(&p, productID)
		if err != nil {
			log.Fatalf("session.Load() failed with %s\n", err)
		}

		err = session.DeleteEntity(p)
		if err != nil {
			log.Fatalf("session.Delete() failed with %s\n", err)
		}

		err = session.SaveChanges()
		if err != nil {
			log.Fatalf("session.SaveChanges() failed with %s\n", err)
		}

		session.Close()
	}

	// verify entity was deleted
	{
		session, err := store.OpenSession("")
		if err != nil {
			log.Fatalf("store.OpenSession() failed with %s\n", err)
		}

		// try to load deleted entity from the server
		var p *northwind.Product
		err = session.Load(&p, productID)
		if err != nil {
			log.Fatalf("session.Load() failed with %s\n", err)
		}

		if p == nil {
			fmt.Printf("Success: we deleted Product with this id so we get nil when we try to load it\n")
		} else {
			fmt.Printf("Error: this entity was deleted so we shouldn't be able to load it\n")
		}

		session.Close()
	}
}

// shows how to query collection of a given name
func queryCollectionByName() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	q := session.QueryCollection("employees")

	var employees []*northwind.Employee
	err = q.GetResults(&employees)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(employees))
}

// shows how to query a collection for a given type
func queryCollectionByType() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)

	var employees []*northwind.Employee
	err = q.GetResults(&employees)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(employees))
}

func queryIndex() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	q := session.QueryIndex("Orders/ByCompany")

	// we're using anonymous struct whose definition matches
	// the fields of in the index
	var ordersByCompany []*struct {
		Company    string
		Count      int
		TotalValue float64 `json:"Total"`
	}
	err = q.GetResults(&ordersByCompany)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(ordersByCompany))
	fmt.Print("First result:\n")
	spew.Fdump(os.Stdout, ordersByCompany[0])
}

// shows how to use First() to get first result
func queryFirst() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)

	var first *northwind.Employee
	err = q.First(&first)
	if err != nil {
		log.Fatalf("q.First() failed with '%s'\n", err)
	}
	// if there are no matching results, first will be unchanged (i.e. equal to nil)
	if first != nil {
		fmt.Print("First() returned:\n")
		spew.Fdump(os.Stdout, first)
	}
}

func queryComplex() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	tp := reflect.TypeOf(&northwind.Product{})
	q := session.QueryCollectionForType(tp)
	q = q.WaitForNonStaleResults(0)
	q = q.WhereEquals("Name", "iPhone X")
	q = q.OrderBy("PricePerUnit")
	q = q.Take(2) // limit to 2 results

	var products []*northwind.Product
	err = q.GetResults(&products)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(products))
}

func querySelectSingleField() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent: from employees select FirstName
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.SelectFields(reflect.TypeOf(""), "FirstName")

	var names []string
	err = q.GetResults(&names)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(names))
	if len(names) > 0 {
		fmt.Printf("First name: %s\n", names[0])
	}
}

type employeeNameTitle struct {
	FirstName string
	Title     string
}

func querySelectFields() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent: from employees select FirstName, Title
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.SelectFields(reflect.TypeOf(&employeeNameTitle{}), "FirstName", "Title")

	var results []*employeeNameTitle
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		spew.Fdump(os.Stdout, results[0])
	}
}

func queryDistinct() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent: from employees select distinct Title
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.SelectFields(reflect.TypeOf(""), "Title")
	q = q.Distinct()

	var results []string
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results. Results: %#v\n", len(results), results)
}

func queryEquals() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent: from employees where Title = 'Sales Representative'
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.WhereEquals("Title", "Sales Representative")

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		spew.Fdump(os.Stdout, results[0])
	}
}

func queryIn() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent: from employees where Title in ['Sales Representative', 'Sales Manager']
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.WhereIn("Title", []interface{}{"Sales Representative", "Sales Manager"})

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		spew.Fdump(os.Stdout, results[0])
	}
}

func main() {
	// to test a given function, uncomment it

	//loadUpdateSave()

	//crudStore()

	//crudLoad()

	//crudLoadWithIncludes()
	// crudUpdate()

	//crudDeleteUsingID()
	//crudDeleteUsingEntity()

	//queryCollectionByName()
	//queryCollectionByType()
	//queryIndex()
	// queryFirst()

	//queryComplex()

	///querySelectSingleField()

	// querySelectFields()
	// queryDistinct()
	// queryEquals()
	queryIn()
}
