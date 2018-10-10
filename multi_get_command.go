package ravendb

import "net/http"

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

	uri := c._baseUrl + "/multi_get"

	return NewHttpPost(uri, nil)
}

/*
       ObjectMapper mapper = JsonExtensions.getDefaultMapper();

       request.setEntity(new ContentProviderHttpEntity(outputStream -> {
           try (JsonGenerator generator = mapper.getFactory().createGenerator(outputStream)) {

               generator.writeStartObject();

               generator.writeFieldName("Requests");
               generator.writeStartArray();

               for (GetRequest command : _commands) {
                   String cacheKey = getCacheKey(command, new Reference<>());

                   Reference<String> cachedChangeVector = new Reference<>();
                   try (CleanCloseable item = _cache.get(cacheKey, cachedChangeVector, new Reference<>())) {
                       Map<String, String> headers = new HashMap<>();
                       if (cachedChangeVector.value != null) {
                           headers.put("If-None-Match", "\"" + cachedChangeVector.value + "\"");
                       }

                       for (Map.Entry<String, String> header : command.getHeaders().entrySet()) {
                           headers.put(header.getKey(), header.getValue());
                       }

                       generator.writeStartObject();

                       generator.writeStringField("Url", "/databases/" + node.getDatabase() + command.getUrl());
                       generator.writeStringField("Query", command.getQuery());

                       generator.writeStringField("Method", command.getMethod());

                       generator.writeFieldName("Headers");
                       generator.writeStartObject();

                       for (Map.Entry<String, String> kvp : headers.entrySet()) {
                           generator.writeStringField(kvp.getKey(), kvp.getValue());
                       }
                       generator.writeEndObject();

                       generator.writeFieldName("Content");
                       if (command.getContent() != null) {
                           command.getContent().writeContent(generator);
                       } else {
                           generator.writeNull();
                       }

                       generator.writeEndObject();
                   }
               }
               generator.writeEndArray();
               generator.writeEndObject();
           } catch (IOException e) {
               throw new RuntimeException(e);
           }
       }, ContentType.APPLICATION_JSON));

       url.value = _baseUrl + "/multi_get";
       return request;
   }
*/

/*
   private String getCacheKey(GetRequest command, Reference<String> requestUrl) {
       requestUrl.value = _baseUrl + command.getUrlAndQuery();
       return command.getMethod() + "-" + requestUrl.value;
   }
*/

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
