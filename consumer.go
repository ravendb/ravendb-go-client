package ravendb

// Go port of https://docs.oracle.com/javase/8/docs/api/java/util/function/Consumer.html

type Consumer interface {
	accept(interface{})
}

type ConsumerOfIndexQuery func(*IndexQuery)
type ConsumerOfQueryResult func(*QueryResult)
type ConsumerOfObjectNode func(ObjectNode)
