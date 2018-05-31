package ravendb

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type GetDatabaseNamesOperation struct {
	_start    int
	_pageSize int
}

func NewGetDatabaseNamesOperation(_start int, _pageSize int) *GetDatabaseNamesOperation {
	return &GetDatabaseNamesOperation{
		_start:    _start,
		_pageSize: _pageSize,
	}
}

func (o *GetDatabaseNamesOperation) getCommand(conventions *DocumentConventions) *RavenCommand {
	return NewGetDatabaseNamesCommand(o._start, o._pageSize)
}

type GetDatabaseNamesCommandData struct {
	_start    int
	_pageSize int
}

func GetDatabaseNamesCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, string) {
	data := cmd.data.(*GetDatabaseNamesCommandData)
	url := node.getUrl() + "/databases?start=" + strconv.Itoa(data._start) + "&pageSize=" + strconv.Itoa(data._pageSize) + "&namesOnly=true"

	return NewHttpGet(), url
}

// GetDatabaseNamesResponse describes response of GetDatabaseNames command
type GetDatabaseNamesResponse struct {
	Databases []string `json:"Databases"`
}

func GetDatabaseNamesCommand_setResponse(cmd *RavenCommand, response String, fromCache bool) error {
	if response == "" {
		return throwInvalidResponse()
	}

	var res GetDatabaseNamesResponse
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}

	/*
		JsonNode names = mapper.readTree(response);
		if (!names.has("Databases")) {
			throwInvalidResponse();
		}

		JsonNode databases = names.get("Databases");
		if (!databases.isArray()) {
			return throwInvalidResponse();
		}
		ArrayNode dbNames = (ArrayNode) databases;
		String[] databaseNames = new String[dbNames.size()];
		for (int i = 0; i < dbNames.size(); i++) {
			databaseNames[i] = dbNames.get(i).asText();
		}
	*/

	cmd.result = res.Databases
	return nil
}

func NewGetDatabaseNamesCommand(_start int, _pageSize int) *RavenCommand {
	data := &GetDatabaseNamesCommandData{
		_start:    _start,
		_pageSize: _pageSize,
	}
	cmd := NewRavenCommand()
	cmd.data = data
	cmd.IsReadRequest = true
	cmd.createRequestFunc = GetDatabaseNamesCommand_createRequest
	cmd.setResponseFunc = GetDatabaseNamesCommand_setResponse
	return cmd
}
