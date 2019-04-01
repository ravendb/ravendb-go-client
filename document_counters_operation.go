package ravendb

type DocumentCountersOperation struct {
	operations []*CounterOperation
	documentId string
}

func (o *DocumentCountersOperation) serialize(conventions *DocumentConventions) map[string]interface{} {
	res := map[string]interface{}{}

	res["DocumentId"] = o.documentId
	if len(o.operations) > 0 {
		ops := []map[string]interface{}{}
		for _, operation := range o.operations {
			ops = append(ops, operation.serialize(conventions))
		}
		res["Operations"] = ops
	}
	return res
}
