package ravendb

/*
public class JsonExtensions {

    private static ObjectMapper _defaultMapper;

    public static ObjectMapper getDefaultMapper() {
        if (_defaultMapper == null) {
            synchronized (JsonExtensions.class) {
                if (_defaultMapper == null) {
                    _defaultMapper = createDefaultJsonSerializer();
                }
            }
        }

        return _defaultMapper;
    }

    public static ObjectMapper createDefaultJsonSerializer() {
        ObjectMapper objectMapper = new ObjectMapper();
        objectMapper.setPropertyNamingStrategy(new DotNetNamingStrategy());
        objectMapper.disable(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES);
        objectMapper.configure(JsonParser.Feature.ALLOW_SINGLE_QUOTES, true);
        objectMapper.setConfig(objectMapper.getSerializationConfig().with(new NetDateFormat()));
        objectMapper.setConfig(objectMapper.getDeserializationConfig().with(new NetDateFormat()));
        objectMapper.setAnnotationIntrospector(new SharpAwareJacksonAnnotationIntrospector());
        return objectMapper;
    }

    public static ObjectMapper getDefaultEntityMapper() {
        ObjectMapper objectMapper = new ObjectMapper();
        objectMapper.disable(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES);
        objectMapper.configure(JsonParser.Feature.ALLOW_SINGLE_QUOTES, true);
        objectMapper.setConfig(objectMapper.getSerializationConfig().with(new NetDateFormat()));
        objectMapper.setConfig(objectMapper.getDeserializationConfig().with(new NetDateFormat()));
        objectMapper.setAnnotationIntrospector(new SharpAwareJacksonAnnotationIntrospector());
        return objectMapper;
    }

    public static class DotNetNamingStrategy extends PropertyNamingStrategy {

        @Override
        public string nameForField(MapperConfig<?> config, AnnotatedField field, string defaultName) {
            return StringUtils.capitalize(defaultName);
        }

        @Override
        public string nameForGetterMethod(MapperConfig<?> config, AnnotatedMethod method, string defaultName) {
            return StringUtils.capitalize(defaultName);
        }

        @Override
        public string nameForSetterMethod(MapperConfig<?> config, AnnotatedMethod method, string defaultName) {
            return StringUtils.capitalize(defaultName);
        }

        @Override
        public string nameForConstructorParameter(MapperConfig<?> config, AnnotatedParameter ctorParam, string defaultName) {
            return StringUtils.capitalize(defaultName);
        }
    }

    public static void writeIndexQuery(JsonGenerator generator, DocumentConventions conventions, IndexQuery query) throws IOException {
        generator.writeStartObject();

        generator.writeStringField("Query", query.getQuery());

        if (query.isPageSizeSet() && query.getPageSize() >= 0) {
            generator.writeNumberField("PageSize", query.getPageSize());
        }

        if (query.isWaitForNonStaleResults()) {
            generator.writeBooleanField("WaitForNonStaleResults", query.isWaitForNonStaleResults());
        }

        if (query.getStart() > 0) {
            generator.writeNumberField("Start", query.getStart());
        }

        if (query.getWaitForNonStaleResultsTimeout() != null) {
            generator.writeStringField("WaitForNonStaleResultsTimeout", TimeUtils.durationToTimeSpan(query.getWaitForNonStaleResultsTimeout()));
        }

        if (query.isDisableCaching()) {
            generator.writeBooleanField("DisableCaching", query.isDisableCaching());
        }

        if (query.isSkipDuplicateChecking()) {
            generator.writeBooleanField("SkipDuplicateChecking", query.isSkipDuplicateChecking());
        }

        generator.writeFieldName("QueryParameters");
        if (query.getQueryParameters() != null) {
            generator.writeObject(EntityToJson.convertEntityToJson(query.getQueryParameters(), conventions));
        } else {
            generator.writeNull();
        }

        generator.writeEndObject();
    }

}
*/

func JsonExtensions_tryGetConflict(metadata ObjectNode) bool {
	v, ok := metadata[Constants_Documents_Metadata_CONFLICT]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		return false
	}
	return b
}
