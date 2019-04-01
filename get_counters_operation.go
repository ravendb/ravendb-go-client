package ravendb

import (
	"errors"
	"net/http"
)

var _ IMaintenanceOperation = &GetCountersOperation{}

type GetCountersOperation struct {
	_docId             string
	_counters          []string
	_returnFullResults bool

	Command *GetCounterValuesCommand
}

func NewGetCountersOperation(docId string, counters []string, returnFullResults bool) *GetCountersOperation {
	return &GetCountersOperation{
		_docId:             docId,
		_counters:          counters,
		_returnFullResults: returnFullResults,
	}
}

func (o *GetCountersOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	var err error
	o.Command, err = NewGetCounterValuesCommand(o._docId, o._counters, o._returnFullResults, conventions)
	if err != nil {
		return nil, err
	}
	return o.Command, nil
}

var (
	_ RavenCommand = &GetCounterValuesCommand{}
)

type GetCounterValuesCommand struct {
	RavenCommandBase

	_docId             string
	_counters          []string
	_returnFullResults bool
	_conventions       *DocumentConventions

	Result *CountersDetail
}

func NewGetCounterValuesCommand(docId string, counters []string, returnFullResults bool, conventions *DocumentConventions) (*GetCounterValuesCommand, error) {
	if docId == "" {
		return nil, newIllegalArgumentError("docId cannot be empty")
	}

	res := &GetCounterValuesCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_docId:             docId,
		_counters:          counters,
		_returnFullResults: returnFullResults,

		_conventions: conventions,
	}

	res.IsReadRequest = true

	return res, nil
}

func (c *GetCounterValuesCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/counters/docId="
	url += urlUtilsEscapeDataString(c._docId)

	if c._returnFullResults {
		url += "&full=true"
	}

	if len(c._counters) > 0 {
		if len(c._counters) > 1 {
			return c.prepareRequestWithMultipleCounters(url)
		} else {
			url += "&counter=" + urlUtilsEscapeDataString(c._counters[0])
		}
	}
	return newHttpGet(url)
}

func (c *GetCounterValuesCommand) prepareRequestWithMultipleCounters(url string) (*http.Request, error) {
	panic("NYI")
	/*
	   HashSet<String> uniqueNames = Sets.newHashSet(_counters);

	   if (uniqueNames.stream().map(x -> x.length()).reduce((a, b) -> a + b).get() < 1024) {
	       for (String uniqueName : uniqueNames) {
	           if (uniqueName != null) {
	               pathBuilder.append("&counter=")
	                       .append(UrlUtils.escapeDataString(uniqueName));
	           } else {
	               HttpPost postRequest = new HttpPost();
	               request = postRequest;

	               DocumentCountersOperation docOps = new DocumentCountersOperation();
	               docOps.setDocumentId(_docId);
	               docOps.setOperations(new ArrayList<>());

	               for (String counter : _counters) {
	                   CounterOperation counterOperation = new CounterOperation();
	                   counterOperation.setType(CounterOperationType.GET);
	                   counterOperation.setCounterName(counter);

	                   docOps.getOperations().add(counterOperation);
	               }

	               CounterBatch batch = new CounterBatch();
	               batch.setDocuments(Arrays.asList(docOps));

	               postRequest.setEntity(new ContentProviderHttpEntity(outputStream -> {
	                   try (JsonGenerator generator = mapper.getFactory().createGenerator(outputStream)) {
	                       batch.serialize(generator, _conventions);
	                   } catch (IOException e) {
	                       throw new RuntimeException(e);
	                   }
	               }, ContentType.APPLICATION_JSON));
	           }
	       }
	   }
	*/
	return nil, errors.New("NYI")
}

func (c *GetCounterValuesCommand) setResponse(response []byte, fromCache bool) error {
	if response == nil {
		return nil
	}

	return jsonUnmarshal(response, &c.Result)
}
