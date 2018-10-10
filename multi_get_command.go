package ravendb

import (
	"encoding/json"
	"net/http"
)

// var _ RavenCommand = &MultiGetCommand{}

type MultiGetCommand struct {
	RavenCommandBase

	data      []*GetResponse // in Java we inherit from List<GetResponse>
	_cache    *HttpCache
	_commands []*GetRequest

	_baseUrl string
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
			v["Method"] = command.method
			v["Headers"] = headers
			v["Content"] = command.content

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

/*
   public void setResponseRaw(CloseableHttpResponse response, InputStream stream) {
       try (JsonParser parser = mapper.getFactory().createParser(stream)) {
           if (parser.nextToken() != JsonToken.START_OBJECT) {
               throwInvalidResponse();
           }

           String property = parser.nextFieldName();
           if (!"Results".equals(property)) {
               throwInvalidResponse();
           }

           int i = 0;
           result = new ArrayList<>();

           for (GetResponse getResponse : readResponses(mapper, parser)) {
               GetRequest command = _commands.get(i);
               maybeSetCache(getResponse, command);
               maybeReadFromCache(getResponse, command);

               result.add(getResponse);

               i++;
           }

           if (parser.nextToken() != JsonToken.END_OBJECT) {
               throwInvalidResponse();
           }


       } catch (Exception e) {
           throwInvalidResponse(e);
       }
   }
*/

/*
   private static List<GetResponse> readResponses(ObjectMapper mapper, JsonParser parser) throws IOException {
       if (parser.nextToken() != JsonToken.START_ARRAY) {
           throwInvalidResponse();
       }

       List<GetResponse> responses = new ArrayList<>();

       while (true) {
           if (parser.nextToken() == JsonToken.END_ARRAY) {
               break;
           }

           responses.add(readResponse(mapper, parser));
       }

       return responses;
   }
*/

/*
   private static GetResponse readResponse(ObjectMapper mapper, JsonParser parser) throws IOException {
       if (parser.currentToken() != JsonToken.START_OBJECT) {
           throwInvalidResponse();
       }

       GetResponse getResponse = new GetResponse();

       while (true) {
           if (parser.nextToken() == null) {
               throwInvalidResponse();
           }

           if (parser.currentToken() == JsonToken.END_OBJECT) {
               break;
           }

           if (parser.currentToken() != JsonToken.FIELD_NAME) {
               throwInvalidResponse();
           }

           String property = parser.getValueAsString();
           switch (property) {
               case "Result":
                   JsonToken jsonToken = parser.nextToken();
                   if (jsonToken == null) {
                       throwInvalidResponse();
                   }

                   if (parser.currentToken() == JsonToken.VALUE_NULL) {
                       continue;
                   }

                   if (parser.currentToken() != JsonToken.START_OBJECT) {
                       throwInvalidResponse();
                   }

                   TreeNode treeNode = mapper.readTree(parser);
                   getResponse.setResult(treeNode.toString());
                   continue;
               case "Headers":
                   if (parser.nextToken() == null) {
                       throwInvalidResponse();
                   }

                   if (parser.currentToken() == JsonToken.VALUE_NULL) {
                       continue;
                   }

                   if (parser.currentToken() != JsonToken.START_OBJECT) {
                       throwInvalidResponse();
                   }

                   ObjectNode headersMap = mapper.readTree(parser);
                   headersMap.fieldNames().forEachRemaining(field -> getResponse.getHeaders().put(field, headersMap.get(field).asText()));
                   continue;
               case "StatusCode":
                   int statusCode = parser.nextIntValue(-1);
                   if (statusCode == -1) {
                       throwInvalidResponse();
                   }

                   getResponse.setStatusCode(statusCode);
                   continue;
               default:
                   throwInvalidResponse();
                   break;

           }


       }

       return getResponse;
   }
*/

/*
   private void maybeReadFromCache(GetResponse getResponse, GetRequest command) {
       if (getResponse.getStatusCode() != HttpStatus.SC_NOT_MODIFIED) {
           return;
       }

       String cacheKey = getCacheKey(command, new Reference<>());
       Reference<String> cachedResponse = new Reference<>();
       try (CleanCloseable cacheItem = _cache.get(cacheKey, new Reference<>(), cachedResponse)) {
           getResponse.setResult(cachedResponse.value);
       }
   }
*/

/*
   private void maybeSetCache(GetResponse getResponse, GetRequest command) {
       if (getResponse.getStatusCode() == HttpStatus.SC_NOT_MODIFIED) {
           return;
       }

       String cacheKey = getCacheKey(command, new Reference<>());

       String result = getResponse.getResult();
       if (result == null) {
           return;
       }

       String changeVector = HttpExtensions.getEtagHeader(getResponse.getHeaders());
       if (changeVector == null) {
           return;
       }

       _cache.set(cacheKey, changeVector, result);
   }
*/
