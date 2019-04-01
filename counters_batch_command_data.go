package ravendb

var _ ICommandData = &CountersBatchCommandData{}

type CountersBatchCommandData struct {
	CommandData

	fromEtl  bool
	counters *DocumentCountersOperation
}

func NewCountersBatchCommandData(documentId string, counterOperations []*CounterOperation) (*CountersBatchCommandData, error) {
	if stringIsWhitespace(documentId) {
		return nil, newIllegalArgumentError("DocumentId cannot be empty")
	}

	counters := &DocumentCountersOperation{
		documentId: documentId,
		operations: counterOperations,
	}
	res := &CountersBatchCommandData{
		CommandData: CommandData{
			ID:   documentId,
			Type: CommandCounters,
		},
		counters: counters,
	}
	return res, nil
}

func (d *CountersBatchCommandData) hasDelete(counterName string) bool {
	return d.hasOperationType(CounterOperationType_DELETE, counterName)
}

func (d *CountersBatchCommandData) hasIncrement(counterName string) bool {
	return d.hasOperationType(CounterOperationType_INCREMENT, counterName)
}

func (d *CountersBatchCommandData) hasOperationType(typ CounterOperationType, counterName string) bool {
	for _, op := range d.counters.operations {
		if counterName != op.counterName {
			continue
		}
		if op.typ == typ {
			return true
		}
	}
	return false
}

func (d *CountersBatchCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	js := d.baseJSON()
	if d.Name != "" {
		js["Name"] = d.Name
	}
	js["Counters"] = d.counters.serialize(conventions)
	js["FromEtl"] = d.fromEtl
	return js, nil
}
