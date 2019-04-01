package ravendb

type CounterBatch struct {
	replyWithAllNodesValues bool
	documents               []*DocumentCountersOperation
	fromEtl                 bool
}

func (c *CounterBatch) serialize(conventions *DocumentConventions) map[string]interface{} {
	res := map[string]interface{}{}
	res["ReplyWithAllNodesValues"] = c.replyWithAllNodesValues
	var docs []map[string]interface{}
	for _, doc := range c.documents {
		js := doc.serialize(conventions)
		docs = append(docs, js)
	}
	res["Documents"] = docs
	res["FromEtl"] = c.fromEtl
	return res
}
