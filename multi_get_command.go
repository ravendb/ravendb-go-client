package ravendb

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

var _ RavenCommand = &MultiGetCommand{}

type MultiGetCommand struct {
	RavenCommandBase

	_cache    *HttpCache
	_commands []*GetRequest
	_baseUrl  string

	Result []*GetResponse // in Java we inherit from List<GetResponse>
}

func NewMultiGetCommand(cache *HttpCache, commands []*GetRequest) *MultiGetCommand {

	cmd := &MultiGetCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_cache:    cache,
		_commands: commands,
	}
	cmd.ResponseType = RavenCommandResponseType_RAW
	return cmd
}

func (c *MultiGetCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	c._baseUrl = node.GetUrl() + "/databases/" + node.GetDatabase()

	m := map[string]interface{}{}
	var requests []map[string]interface{}

	for _, command := range c._commands {
		v := map[string]interface{}{}
		cacheKey, _ := c.getCacheKey(command)
		{
			item, cachedChangeVector, _ := c._cache.get(cacheKey)
			headers := map[string]string{}
			if cachedChangeVector != nil {
				headers["If-None-Match"] = "\"" + *cachedChangeVector + "\""
			}
			for k, v := range command.headers {
				headers[k] = v
			}
			v["Url"] = "/databases/" + node.GetDatabase() + command.url
			v["Query"] = command.query
			if command.method == "" {
				v["Method"] = nil
			} else {
				v["Method"] = command.method
			}
			v["Headers"] = headers
			if command.content != nil {
				v["Content"] = command.content.writeContent()
			} else {
				v["Content"] = nil
			}

			item.Close()
		}
		requests = append(requests, v)
	}

	m["Requests"] = requests
	d, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	uri := c._baseUrl + "/multi_get"
	return NewHttpPost(uri, d)
}

func (c *MultiGetCommand) getCacheKey(command *GetRequest) (string, string) {
	uri := c._baseUrl + command.getUrlAndQuery()
	key := command.method + "-" + uri
	return key, uri
}

type getResponseJSON struct {
	Result     json.RawMessage   `json:"Result"`
	StatusCode int               `json:"StatusCode"`
	Headers    map[string]string `json:"Headers"`
}

type resultsJSON struct {
	Results []*getResponseJSON `json:"Results"`
}

func (c *MultiGetCommand) SetResponseRaw(response *http.Response, stream io.Reader) error {
	var results *resultsJSON
	d, err := ioutil.ReadAll(stream)
	if err != nil {
		return err
	}
	err = json.Unmarshal(d, &results)
	if err != nil {
		return err
	}

	for i, rsp := range results.Results {
		command := c._commands[i]
		var getResponse GetResponse

		getResponse.statusCode = rsp.StatusCode
		getResponse.headers = rsp.Headers
		getResponse.result = string(rsp.Result)

		c.maybeSetCache(&getResponse, command)
		c.maybeReadFromCache(&getResponse, command)

		c.Result = append(c.Result, &getResponse)
	}

	return nil
}

func (c *MultiGetCommand) maybeReadFromCache(getResponse *GetResponse, command *GetRequest) {
	if getResponse.statusCode != http.StatusNotModified {
		return
	}

	cacheKey, _ := c.getCacheKey(command)
	{
		cacheItem, _, cachedResponse := c._cache.get(cacheKey)
		getResponse.result = cachedResponse
		cacheItem.Close()
	}
}

func (c *MultiGetCommand) maybeSetCache(getResponse *GetResponse, command *GetRequest) {
	if getResponse.statusCode == http.StatusNotModified {
		return
	}

	cacheKey, _ := c.getCacheKey(command)

	result := getResponse.result
	if result == "" {
		return
	}

	changeVector := HttpExtensions_getEtagHeaderFromMap(getResponse.headers)
	if changeVector == nil {
		return
	}

	c._cache.set(cacheKey, changeVector, result)
}
