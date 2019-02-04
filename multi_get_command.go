package ravendb

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

var _ RavenCommand = &MultiGetCommand{}

// MultiGetCommand represents multi get command
type MultiGetCommand struct {
	RavenCommandBase

	cache    *HttpCache
	commands []*GetRequest
	baseURL  string

	Result []*GetResponse // in Java we inherit from List<GetResponse>
}

// NewMultiGetCommand returns new MultiGetCommand
func NewMultiGetCommand(cache *HttpCache, commands []*GetRequest) *MultiGetCommand {

	cmd := &MultiGetCommand{
		RavenCommandBase: NewRavenCommandBase(),

		cache:    cache,
		commands: commands,
	}
	cmd.ResponseType = RavenCommandResponseTypeRaw
	return cmd
}

// CreateRequest creates http request for this command
func (c *MultiGetCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	c.baseURL = node.URL + "/databases/" + node.Database

	m := map[string]interface{}{}
	var requests []map[string]interface{}

	for _, command := range c.commands {
		v := map[string]interface{}{}
		cacheKey, _ := c.getCacheKey(command)
		{
			item, cachedChangeVector, _ := c.cache.get(cacheKey)
			headers := map[string]string{}
			if cachedChangeVector != nil {
				headers[headersIfNoneMatch] = "\"" + *cachedChangeVector + "\""
			}
			for k, v := range command.headers {
				headers[k] = v
			}
			v["Url"] = "/databases/" + node.Database + command.url
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
	d, err := jsonMarshal(m)
	if err != nil {
		return nil, err
	}

	uri := c.baseURL + "/multi_get"
	return NewHttpPost(uri, d)
}

func (c *MultiGetCommand) getCacheKey(command *GetRequest) (string, string) {
	uri := c.baseURL + command.getUrlAndQuery()
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

// SetResponseRaw sets response from http response
func (c *MultiGetCommand) SetResponseRaw(response *http.Response, stream io.Reader) error {
	var results *resultsJSON
	d, err := ioutil.ReadAll(stream)
	if err != nil {
		return err
	}
	err = jsonUnmarshal(d, &results)
	if err != nil {
		return err
	}

	for i, rsp := range results.Results {
		command := c.commands[i]
		var getResponse GetResponse

		getResponse.StatusCode = rsp.StatusCode
		getResponse.Headers = rsp.Headers
		getResponse.Result = rsp.Result

		c.maybeSetCache(&getResponse, command)
		c.maybeReadFromCache(&getResponse, command)

		c.Result = append(c.Result, &getResponse)
	}

	return nil
}

func (c *MultiGetCommand) maybeReadFromCache(getResponse *GetResponse, command *GetRequest) {
	if getResponse.StatusCode != http.StatusNotModified {
		return
	}

	cacheKey, _ := c.getCacheKey(command)
	{
		cacheItem, _, cachedResponse := c.cache.get(cacheKey)
		getResponse.Result = cachedResponse
		cacheItem.Close()
	}
}

func (c *MultiGetCommand) maybeSetCache(getResponse *GetResponse, command *GetRequest) {
	if getResponse.StatusCode == http.StatusNotModified {
		return
	}

	cacheKey, _ := c.getCacheKey(command)

	result := getResponse.Result
	if len(result) == 0 {
		return
	}

	changeVector := gttpExtensionsGetEtagHeaderFromMap(getResponse.Headers)
	if changeVector == nil {
		return
	}

	c.cache.set(cacheKey, changeVector, result)
}
