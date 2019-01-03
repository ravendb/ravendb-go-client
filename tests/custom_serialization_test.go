package tests

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

const (
	Dollar = "USD"
	Euro   = "EUR"
)

func customSerializationTestSerialization(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		product1 := &Product3{
			Name:  "iPhone",
			Price: NewMoney(9999, Dollar),
		}
		product2 := &Product3{
			Name:  "Camera",
			Price: NewMoney(150, Euro),
		}
		product3 := &Product3{
			Name:  "Bread",
			Price: NewMoney(2, Dollar),
		}
		err = session.Store(product1)
		assert.NoError(t, err)
		err = session.Store(product2)
		assert.NoError(t, err)
		err = session.Store(product3)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	// verify if value was properly serialized
	{
		command := ravendb.NewGetDocumentsCommand([]string{"product3s/1-A"}, nil, false)
		err = store.GetRequestExecutor("").ExecuteCommand(command)
		assert.NoError(t, err)
		productJSON := command.Result.Results[0]
		priceNode := productJSON["price"]
		assert.Equal(t, priceNode, "9999 USD")
	}

	//verify if query properly serialize value
	{
		session := openSessionMust(t, store)

		var productsForTwoDollars []*Product3
		q := session.Query()
		q = q.WhereEquals("price", NewMoney(2, Dollar))
		err := q.GetResults(&productsForTwoDollars)
		assert.NoError(t, err)

		assert.Equal(t, len(productsForTwoDollars), 1)

		product := productsForTwoDollars[0]
		assert.Equal(t, product.Name, "Bread")

		session.Close()
	}
}

// unique name to not conflict with Proudct and Product2 elsewhere
type Product3 struct {
	Name  string `json:"name"`
	Price Money  `json:"price"`
}

type Money struct {
	Currency string `json:"currency"`
	Amount   int    `json:"amount"`
}

func NewMoney(n int, curr string) Money {
	return Money{
		Currency: curr,
		Amount:   n,
	}
}

func (m Money) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf(`"%d %s"`, m.Amount, m.Currency)
	return []byte(s), nil
}

func (m *Money) UnmarshalJSON(d []byte) error {
	s := string(d)
	s = strings.TrimPrefix(s, `"`)
	s = strings.TrimSuffix(s, `"`)
	parts := strings.Split(s, " ")
	if len(parts) != 2 {
		return fmt.Errorf("'%s' is not valid JSON serialization for Money", s)
	}
	n, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("'%s' is not valid JSON serialization for Money", s)
	}
	m.Amount = n
	m.Currency = parts[1]
	return nil
}

func TestCustomSerialization(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches the order of Java tests
	customSerializationTestSerialization(t, driver)
}
