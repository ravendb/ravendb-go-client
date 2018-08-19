package ravendb

type ISuggestionBuilder interface {
	ByField(fieldName string, term string, terms ...string) ISuggestionOperations

	//TBD expr ISuggestionOperations<T> ByField(Expression<Func<T, object>> path, string term);
	//TBD expr ISuggestionOperations<T> ByField(Expression<Func<T, object>> path, string[] terms);

}
