package ravendb

type moreLikeThisScope struct {
	token             *moreLikeThisToken
	addQueryParameter func(interface{}) string
	onDispose         func()
}

func newMoreLikeThisScope(token *moreLikeThisToken, addQueryParameter func(interface{}) string, onDispose func()) *moreLikeThisScope {
	return &moreLikeThisScope{
		token:             token,
		addQueryParameter: addQueryParameter,
		onDispose:         onDispose,
	}
}

func (s *moreLikeThisScope) Close() {
	if s.onDispose != nil {
		s.onDispose()
	}
}

func (s *moreLikeThisScope) withOptions(options *MoreLikeThisOptions) {
	if options == nil {
		return
	}

	// force using *non* entity serializer here:
	optionsAsJson := valueToTree(options)
	s.token.optionsParameterName = s.addQueryParameter(optionsAsJson)
}

func (s *moreLikeThisScope) withDocument(document string) {
	s.token.documentParameterName = s.addQueryParameter(document)
}
