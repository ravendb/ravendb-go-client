package ravendb

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type _GetDocumentsCommand struct {
	_id String

	_ids      []String
	_includes []String

	_metadataOnly bool

	_startWith  String
	_matches    String
	_start      int
	_pageSize   int
	_exclude    String
	_startAfter String
}

func NewGetDocumentsCommand(ids []string, includes []string, metadataOnly bool) *RavenCommand {
	d := &_GetDocumentsCommand{
		_includes:     includes,
		_metadataOnly: metadataOnly,
	}

	if len(ids) == 1 {
		d._id = ids[0]
	} else {
		d._ids = ids
	}
	return NewGetDocumentsCommandWithData(d)
}

func NewGetDocumentsCommandWithData(data *_GetDocumentsCommand) *RavenCommand {
	cmd := NewRavenCommand()
	cmd.data = data
	cmd.IsReadRequest = true
	cmd.createRequestFunc = GetDocumentsCommand_createRequest
	cmd.setResponseFunc = GetDocumentsCommand_setResponse
	return cmd
}

func GetDocumentsCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, string) {
	data := cmd.data.(*_GetDocumentsCommand)
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/docs?"
	// TODO: is _start == 0 valid?
	if data._start != 0 {
		url += "&start=" + strconv.Itoa(data._start)
	}

	if data._pageSize != 0 {
		url += "&pageSize=" + strconv.Itoa(data._pageSize)
	}

	if data._metadataOnly {
		url += "&metadataOnly=true"
	}

	// TODO: more

	return NewHttpGet(url), url
}

func GetDocumentsCommand_setResponse(cmd *RavenCommand, response String, fromCache bool) error {
	if response == "" {
		return nil
	}

	var res GetDocumentsResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	cmd.result = &res
	return nil
}
