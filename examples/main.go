package main

import (
	"bytes"
	"fmt"
	"github.com/ravendb/ravendb-go-client/examples/northwind"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/kylelemons/godebug/pretty"
	"github.com/ravendb/ravendb-go-client"
)

// "Demo" is a Northwind sample database
// You can browse its content via web interface at
// http://live-test.ravendb.net/studio/index.html#databases/documents?&database=Demo
var (
	dbName = "Demo"
)

func printRQL(q *ravendb.DocumentQuery) {
	iq, err := q.GetIndexQuery()
	if err != nil {
		log.Fatalf("q.GetIndexQuery() returned '%s'\n", err)
	}
	fmt.Printf("RQL: %s\n", iq.GetQuery())
	params := iq.GetQueryParameters()
	if len(params) == 0 {
		return
	}
	fmt.Printf("Parameters:\n")
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("  $%s: %#v\n", key, params[key])
	}
	fmt.Print("\n")
}

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
		Category:     "electronics",
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
	fmt.Print("empolyee:\n")
	pretty.Print(e)
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
			Category:     "electronics",
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
			Category:     "electronics",
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

		err = session.DeleteByID(productID, nil)
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
			Category:     "electronics",
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

		err = session.Delete(p)
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
	printRQL(q)

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
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
	printRQL(q)

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func queryIndex() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	q := session.QueryIndex("Orders/ByCompany")
	printRQL(q)

	// we're using anonymous struct whose definition matches
	// the fields of in the index
	var results []*struct {
		Company    string
		Count      int
		TotalValue float64 `json:"Total"`
	}
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
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
	printRQL(q)

	var first *northwind.Employee
	err = q.First(&first)
	if err != nil {
		log.Fatalf("q.First() failed with '%s'\n", err)
	}
	// if there are no matching results, first will be unchanged (i.e. equal to nil)
	if first != nil {
		fmt.Print("First() returned:\n")
		pretty.Print(first)
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
	printRQL(q)

	var results []*northwind.Product
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func querySelectSingleField() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent:
	// from employees select FirstName
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.SelectFields(reflect.TypeOf(""), "FirstName")
	printRQL(q)

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

	// RQL equivalent:
	// from employees select FirstName, Title
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.SelectFields(reflect.TypeOf(&employeeNameTitle{}), "FirstName", "Title")
	printRQL(q)

	var results []*employeeNameTitle
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func queryDistinct() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent:
	// from employees select distinct Title
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.SelectFields(reflect.TypeOf(""), "Title")
	q = q.Distinct()
	printRQL(q)

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

	// RQL equivalent:
	// from employees where Title = 'Sales Representative'
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.WhereEquals("Title", "Sales Representative")
	printRQL(q)

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func queryIn() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent:
	// from employees where Title in ['Sales Representative', 'Sales Manager']
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.WhereIn("Title", []interface{}{"Sales Representative", "Sales Manager"})
	printRQL(q)

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func queryStartsWith() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent:
	// from employees where startsWith('Ro')
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.WhereStartsWith("FirstName", "Ro")
	printRQL(q)

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func queryEndsWith() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent:
	// from employees where endsWith('rt')
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.WhereEndsWith("FirstName", "rt")
	printRQL(q)

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func queryBetween() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent:
	// from orders where Freight between 11 and 13
	tp := reflect.TypeOf(&northwind.Order{})
	q := session.QueryCollectionForType(tp)
	q = q.WhereBetween("Freight", 11, 13)
	printRQL(q)

	var results []*northwind.Order
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func queryGreater() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent:
	// from orders where Freight Freight > 11
	tp := reflect.TypeOf(&northwind.Order{})
	q := session.QueryCollectionForType(tp)
	// can also be WhereGreaterThanOrEqual(), WhereLessThan(), WhereLessThanOrEqual()
	q = q.WhereGreaterThan("Freight", 11)
	printRQL(q)

	var results []*northwind.Order
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func queryExists() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent:
	// from employees where exists ("ReportsTo")
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.WhereExists("ReportsTo")
	printRQL(q)

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func queryContainsAny() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent:
	// from employees where FirstName in ("Anne", "Nancy")
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	// can also be ContainsAll()
	q = q.ContainsAny("FirstName", []interface{}{"Anne", "Nancy"})
	printRQL(q)

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func querySearch() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent:
	// from employees where search(FirstName, 'Anne Nancy')
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.Search("FirstName", "Anne Nancy")
	printRQL(q)

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func querySubclause() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent:
	// from employees where (FirstName = 'Steven') or (Title = 'Sales Representative' and LastName = 'Davolio')
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.WhereEquals("FirstName", "Steven")
	q = q.OrElse()
	q = q.OpenSubclause()
	q = q.WhereEquals("Title", "Sales Representative")
	q = q.WhereEquals("LastName", "Davolio")
	q = q.CloseSubclause()
	printRQL(q)

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func queryNot() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent:
	// from employees where not FirstName = 'Steven'
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.Not()
	q = q.WhereEquals("FirstName", "Steven")
	printRQL(q)

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func queryOrElse() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent:
	// from employees where FirstName = 'Steven' or FirstName  = 'Nancy'
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.WhereEquals("FirstName", "Steven")
	// can also be AndElse()
	q = q.OrElse()
	q = q.WhereEquals("FirstName", "Nancy")
	printRQL(q)

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func queryOrderBy() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent:
	// from employees order by FirstName
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	// can also be RandomOrdering()
	q = q.OrderBy("FirstName")
	printRQL(q)

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func queryTake() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent:
	// from employees order by FirstName desc
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.OrderByDescending("FirstName")
	q = q.Take(2)
	printRQL(q)

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func querySkip() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// RQL equivalent:
	// from employees order by FirstName desc
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.OrderByDescending("FirstName")
	q = q.Take(2)
	q = q.Skip(1)
	printRQL(q)

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	if len(results) > 0 {
		fmt.Print("First result:\n")
		pretty.Print(results[0])
	}
}

func queryStatistics() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	var stats *ravendb.QueryStatistics
	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.WhereGreaterThan("FirstName", "Bernard")
	q = q.OrderByDescending("FirstName")
	q.Statistics(&stats)
	printRQL(q)

	var results []*northwind.Employee
	err = q.GetResults(&results)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", len(results))
	fmt.Printf("Statistics:\n")
	pretty.Print(stats)
}

func querySingle() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.WhereEquals("LastName", "Davolio")
	printRQL(q)

	var result *northwind.Employee
	err = q.Single(&result)
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned:\n")
	pretty.Print(result)
}

func queryCount() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	q = q.WhereGreaterThan("LastName", "Davolio")
	printRQL(q)

	n, err := q.Count()
	if err != nil {
		log.Fatalf("q.GetResults() failed with '%s'\n", err)
	}
	fmt.Printf("Query returned %d results\n", n)
}

// auto-detect path to "examples" directory
func dataDir() string {
	dir := "."
	_, err := os.Stat("examples")
	if err == nil {
		dir = "examples"
	}
	path, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("filepath.Abs() failed with '%s'\n", err)
	}
	return path
}

func storeAttachments() string {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	e := &northwind.Employee{
		FirstName: "Jon",
		LastName:  "Snow",
	}
	err = session.Store(e)
	if err != nil {
		log.Fatalf("session.Store() failed with '%s'\n", err)
	}

	path := filepath.Join(dataDir(), "pic.png")
	fileStream, err := os.Open(path)
	if err != nil {
		log.Fatalf("os.Open() failed with '%s'\n", err)
	}
	defer fileStream.Close()

	fmt.Printf("new employee id: %s\n", e.ID)
	err = session.Advanced().Attachments().Store(e, "photo.png", fileStream, "image/png")

	// could also be done using document id
	// err = session.Advanced().Attachments().Store(e.ID, "photo.png", fileStream, "image/png")

	if err != nil {
		log.Fatalf("session.Advanced().Attachments().Store() failed with '%s'\n", err)
	}

	err = session.SaveChanges()
	if err != nil {
		log.Fatalf("session.SaveChanges() failed with '%s'\n", err)
	}

	return e.ID
}

func getAttachments() {
	docID := storeAttachments()
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	attachment, err := session.Advanced().Attachments().GetByID(docID, "photo.png")
	if err != nil {
		log.Fatalf("session.Advanced().Attachments().Get() failed with '%s'\n", err)
	}
	defer attachment.Close()
	fmt.Print("Attachment details:\n")
	pretty.Print(attachment.Details)
	// read attachment data
	// attachment.Data is io.Reader
	var attachmentData bytes.Buffer
	n, err := io.Copy(&attachmentData, attachment.Data)
	if err != nil {
		log.Fatalf("io.Copy() failed with '%s'\n", err)
	}
	fmt.Printf("Attachment size: %d bytes\n", n)
}

func checkAttachmentExists() {
	docID := storeAttachments()
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	{
		name := "photo.png"
		exists, err := session.Advanced().Attachments().Exists(docID, name)
		if err != nil {
			log.Fatalf("session.Advanced().Attachments().Exists() failed with '%s'\n", err)
		}
		if exists {
			fmt.Printf("attachment '%s' exists\n", name)
		} else {
			fmt.Printf("attachment '%s' doesn't exists\n", name)
		}
	}

	{
		name := "non-existent.png"
		exists, err := session.Advanced().Attachments().Exists(docID, name)
		if err != nil {
			log.Fatalf("session.Advanced().Attachments().Exists() failed with '%s'\n", err)
		}
		if exists {
			fmt.Printf("attachment '%s' exists\n", name)
		} else {
			fmt.Printf("attachment '%s' doesn't exists\n", name)
		}
	}
}

func getAttachmentNames() {
	docID := storeAttachments()
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	var doc *northwind.Employee
	err = session.Load(&doc, docID)
	if err != nil {
		log.Fatalf("session.Load() failed with '%s'\n", err)
	}

	names, err := session.Advanced().Attachments().GetNames(doc)
	if err != nil {
		log.Fatalf("session.Advanced().Attachments().GetNames() failed with '%s'\n", err)
	}
	fmt.Print("Attachment names:\n")
	pretty.Print(names)
}

func bulkInsert() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	bulkInsert := store.BulkInsert("")

	names := []string{"Anna", "Maria", "Miguel", "Emanuel", "Dayanara", "Aleida"}
	var ids []string
	for _, name := range names {
		e := &northwind.Employee{
			FirstName: name,
		}
		id, err := bulkInsert.Store(e, nil)
		if err != nil {
			log.Fatalf("bulkInsert.Store() failed with '%s'\n", err)
		}
		ids = append(ids, id)
	}
	// flush data and finish
	err = bulkInsert.Close()
	if err != nil {
		log.Fatalf("bulkInsert.Close() failed with '%s'\n", err)
	}

	fmt.Printf("Finished %d documents with ids: %v\n", len(names), ids)
}

func changes() {
	//ravendb.EnableDatabaseChangesDebugOutput = true

	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	changes := store.Changes("")

	err = changes.EnsureConnectedNow()
	if err != nil {
		log.Fatalf("changes.EnsureConnectedNow() failed with '%s'\n", err)
	}

	var wg sync.WaitGroup
	onDocChange := func(change *ravendb.DocumentChange) {
		fmt.Print("change:\n")
		pretty.Print(change)
		wg.Done()
	}
	docChangesCancel, err := changes.ForAllDocuments(onDocChange)
	if err != nil {
		log.Fatalf("changes.ForAllDocuments() failed with '%s'\n", err)
	}
	defer docChangesCancel()

	wg.Add(1)
	e := &northwind.Employee{
		FirstName: "Jon",
		LastName:  "Snow",
	}
	err = session.Store(e)
	if err != nil {
		log.Fatalf("session.Store() failed with '%s'\n", err)
	}

	err = session.SaveChanges()
	if err != nil {
		log.Fatalf("session.SaveChanges() failed with '%s'\n", err)
	}

	timeStart := time.Now()
	fmt.Print("Waiting for the change\n")
	// wait for the change to be received
	wg.Wait()
	fmt.Printf("Took %s to receive change notifications\n", time.Since(timeStart))
}

func streamWithIDPrefix() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	args := &ravendb.StartsWithArgs{
		StartsWith: "products/",
	}
	iterator, err := session.Advanced().Stream(args)
	if err != nil {
		log.Fatalf("session.Advanced().Stream() failed with '%s'\n", err)
	}
	n := 0
	for {
		var p *northwind.Product
		streamResult, err := iterator.Next(&p)
		if err != nil {
			// io.EOF means there are no more results
			if err == io.EOF {
				err = nil
			} else {
				log.Fatalf("iterator.Next() failed with '%s'\n", err)
			}
			break
		}
		if n < 1 {
			fmt.Print("streamResult:\n")
			pretty.Print(streamResult)
			fmt.Print("product:\n")
			pretty.Print(p)
			fmt.Print("\n")
		}
		n++
	}
	fmt.Printf("Got %d results\n", n)
}

func streamQueryResults() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	tp := reflect.TypeOf(&northwind.Product{})
	q := session.QueryCollectionForType(tp)
	q = q.WhereGreaterThan("PricePerUnit", 15)
	q = q.OrderByDescending("PricePerUnit")

	iterator, err := session.Advanced().StreamQuery(q, nil)
	if err != nil {
		log.Fatalf("session.Advanced().StreamQuery() failed with '%s'\n", err)
	}
	n := 0
	for {
		var p *northwind.Product
		streamResult, err := iterator.Next(&p)
		if err != nil {
			// io.EOF means there are no more results
			if err == io.EOF {
				err = nil
			} else {
				log.Fatalf("iterator.Next() failed with '%s'\n", err)
			}
			break
		}
		if n < 1 {
			fmt.Print("streamResult:\n")
			pretty.Print(streamResult)
			fmt.Print("product:\n")
			pretty.Print(p)
			fmt.Print("\n")
		}
		n++
	}
	fmt.Printf("Got %d results\n", n)
}

func setupRevisions(store *ravendb.DocumentStore, purgeOnDelete bool, minimumRevisionsToKeep int64) (*ravendb.ConfigureRevisionsOperationResult, error) {

	revisionsConfiguration := &ravendb.RevisionsConfiguration{}
	defaultCollection := &ravendb.RevisionsCollectionConfiguration{}
	defaultCollection.PurgeOnDelete = purgeOnDelete
	defaultCollection.MinimumRevisionsToKeep = minimumRevisionsToKeep

	revisionsConfiguration.DefaultConfig = defaultCollection
	operation := ravendb.NewConfigureRevisionsOperation(revisionsConfiguration)

	err := store.Maintenance().Send(operation)
	if err != nil {
		return nil, err
	}

	return operation.Command.Result, nil
}

func revisions() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	// first must configure the store to enable revisions
	_, err = setupRevisions(store, true, 5)
	if err != nil {
		log.Fatalf("setupRevisions() failed with '%s'\n", err)
	}

	e := &northwind.Employee{
		FirstName: "Jon",
		LastName:  "Snow",
	}
	err = session.Store(e)
	if err != nil {
		log.Fatalf("session.Store() failed with '%s'\n", err)
	}
	err = session.SaveChanges()
	if err != nil {
		log.Fatalf("session.SaveChanges() failed with '%s'\n", err)
	}

	// modify document to create a new revision
	e.FirstName = "Jhonny"
	err = session.SaveChanges()
	if err != nil {
		log.Fatalf("session.SaveChanges() failed with '%s'\n", err)
	}

	var revisions []*northwind.Employee
	err = session.Advanced().Revisions().GetFor(&revisions, e.ID)
	if err != nil {
		log.Fatalf(" session.Advanced().Revisions().GetFor() failed with '%s'\n", err)
	}
	pretty.Print(revisions)
}

func suggestions() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	index := ravendb.NewIndexCreationTask("EmployeeIndex")
	index.Map = "from doc in docs.Employees select new { doc.FirstName }"
	index.Suggestion("FirstName")

	err = store.ExecuteIndex(index, "")
	if err != nil {
		log.Fatalf("store.ExecuteIndex() failed with '%s'\n", err)
	}

	tp := reflect.TypeOf(&northwind.Employee{})
	q := session.QueryCollectionForType(tp)
	su := ravendb.NewSuggestionWithTerm("FirstName")
	su.Term = "Micael"
	suggestionQuery := q.SuggestUsing(su)
	results, err := suggestionQuery.Execute()
	if err != nil {
		log.Fatalf("suggestionQuery.Execute() failed with '%s'\n", err)
	}
	pretty.Print(results)
}

func advancedPatching() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	product := &northwind.Product{
		Name:         "iPhone X",
		PricePerUnit: 50,
		Category:     "electronics",
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

	err = session.Advanced().IncrementByID(product.ID, "PricePerUnit", 15)
	if err != nil {
		log.Fatalf("session.Advanced().IncrementByID() failed with %s\n", err)
	}

	err = session.Advanced().Patch(product, "Category", "expensive products")
	if err != nil {
		log.Fatalf("session.Advanced().PatchEntity() failed with %s\n", err)
	}

	err = session.SaveChanges()
	if err != nil {
		log.Fatalf("session.SaveChanges() failed with %s\n", err)
	}

	{
		newSession, err := store.OpenSession("")
		if err != nil {
			log.Fatalf("store.OpenSession() failed with %s\n", err)
		}

		var p *northwind.Product
		err = newSession.Load(&p, product.ID)
		if err != nil {
			log.Fatalf("newSession.Load() failed with %s\n", err)
		}
		pretty.Print(p)

		newSession.Close()
	}
}

func subscriptions() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	opts := ravendb.SubscriptionCreationOptions{
		Query: "from Products where PricePerUnit > 17 and PricePerUnit < 19",
	}
	tp := reflect.TypeOf(&northwind.Product{})
	subscriptionName, err := store.Subscriptions().CreateForType(tp, &opts, "")
	if err != nil {
		log.Fatalf("store.Subscriptions().Create() failed with %s\n", err)
	}
	wopts := ravendb.NewSubscriptionWorkerOptions(subscriptionName)
	worker, err := store.Subscriptions().GetSubscriptionWorker(tp, wopts, "")
	if err != nil {
		log.Fatalf("store.Subscriptions().GetSubscriptionWorker() failed with %s\n", err)
	}

	chResults := make(chan bool, 64)
	processItems := func(batch *ravendb.SubscriptionBatch) error {
		fmt.Print("Batch of subscription results:\n")
		pretty.Print(batch)
		chResults <- true
		return nil
	}
	err = worker.Run(processItems)

	// wait for at least one batch result
	select {
	case <-chResults:
	// no-op
	case <-time.After(time.Second * 5):
		fmt.Printf("Timed out waiting for first subscription batch\n")
	}

	_ = worker.Close()
}

func main() {
	// to test a given function, uncomment it
	//loadUpdateSave()
	//crudStore()
	//crudLoad()
	//crudLoadWithIncludes()
	//crudUpdate()
	//crudDeleteUsingID()
	//crudDeleteUsingEntity()
	//queryCollectionByName()
	//queryCollectionByType()
	//queryIndex()
	//queryFirst()
	//queryComplex()
	//querySelectSingleField()
	//querySelectFields()
	//queryDistinct()
	//queryEquals()
	//queryIn()
	//queryStartsWith()
	//queryEndsWith()
	//queryBetween()
	//queryGreater()
	//queryExists()
	//queryContainsAny()
	//querySearch()
	//querySubclause()
	//queryNot()
	//queryOrElse()
	//queryOrderBy()
	//queryTake()
	//querySkip()
	//queryStatistics()
	//querySingle()
	//queryCount()
	//storeAttachments()
	//getAttachments()
	//checkAttachmentExists()
	//getAttachmentNames()
	//bulkInsert()
	//changes()
	//streamWithIDPrefix()
	//streamQueryResults()
	//revisions()
	//suggestions()
	//advancedPatching()

	subscriptions()
}
