package ravendb

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type _GetDocumentsCommand struct {
	_id string

	_ids      []string
	_includes []string

	_metadataOnly bool

	_startWith  string
	_matches    string
	_start      int
	_pageSize   int
	_exclude    string
	_startAfter string
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

func GetDocumentsCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, error) {
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

	if data._startWith != "" {
		url += "&startsWith="
		url += UrlUtils_escapeDataString(data._startWith)

		if data._matches != "" {
			url += "&matches="
			url += data._matches
		}

		if data._exclude != "" {
			url += "&exclude="
			url += data._exclude
		}

		if data._startAfter != "" {
			url += "&startAfter="
			url += data._startAfter
		}
	}

	for _, include := range data._includes {
		url += "&include="
		url += include
	}

	if data._id != "" {
		url += "&id="
		url += UrlUtils_escapeDataString(data._id)
		return NewHttpGet(url)
	}

	panicIf(len(data._ids) == 0, "must provide _id or _ids")

	return GetDocumentsCommand_prepareRequestWithMultipleIds(url, data._ids)
}

func GetDocumentsCommand_prepareRequestWithMultipleIds(url string, ids []string) (*http.Request, error) {
	panicIf(true, "NYI")
	return nil, errors.New("NYI")
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
