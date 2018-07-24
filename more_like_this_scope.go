package ravendb

type MoreLikeThisScope struct {
	_token             *MoreLikeThisToken
	_addQueryParameter func(Object) string
	_onDispose         func()
}

func NewMoreLikeThisScope(token *MoreLikeThisToken, addQueryParameter func(Object) string, onDispose func()) *MoreLikeThisScope {
	return &MoreLikeThisScope{
		_token:             token,
		_addQueryParameter: addQueryParameter,
		_onDispose:         onDispose,
	}
}

func (s *MoreLikeThisScope) Close() {
	if s._onDispose != nil {
		s._onDispose()
	}
}

func (s *MoreLikeThisScope) withOptions(options *MoreLikeThisOptions) {
	if options == nil {
		return
	}

	// force using *non* entity serializer here:
	optionsAsJson := valueToTree(options)
	s._token.optionsParameterName = s._addQueryParameter(optionsAsJson)
}

func (s *MoreLikeThisScope) withDocument(document string) {
	s._token.documentParameterName = s._addQueryParameter(document)
}
