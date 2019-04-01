package ravendb

import (
	"encoding/json"
	"net/http"
)

var _ IMaintenanceOperation = &CounterBatchOperation{}

type CounterBatchOperation struct {
	_counterBatch *CounterBatch

	Command *CounterBatchCommand
}

func NewCounterBatchOperation(counterBatch *CounterBatch) *CounterBatchOperation {
	return &CounterBatchOperation{
		_counterBatch: counterBatch,
	}
}

func (o *CounterBatchOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	var err error
	o.Command, err = NewCounterBatchCommand(o._counterBatch, conventions)
	if err != nil {
		return nil, err
	}
	return o.Command, nil
}

var (
	_ RavenCommand = &CounterBatchCommand{}
)

type CounterBatchCommand struct {
	RavenCommandBase

	_conventions  *DocumentConventions
	_counterBatch *CounterBatch

	Result *CountersDetail
}

func NewCounterBatchCommand(counterBatch *CounterBatch, conventions *DocumentConventions) (*CounterBatchCommand, error) {
	if counterBatch == nil {
		return nil, newIllegalArgumentError("counterBatch cannot be nil")
	}

	res := &CounterBatchCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_conventions:  conventions,
		_counterBatch: counterBatch,
	}
	return res, nil
}

func (c *CounterBatchCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/counters"

	js := c._counterBatch.serialize(c._conventions)
	d, err := json.Marshal(js)
	if err != nil {
		return nil, err
	}

	return newHttpPost(url, d)
}

func (c *CounterBatchCommand) setResponse(response []byte, fromCache bool) error {
	if response == nil {
		return nil
	}

	return jsonUnmarshal(response, &c.Result)
}
