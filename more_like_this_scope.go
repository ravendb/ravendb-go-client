package ravendb

type MoreLikeThisScope struct {
	_token             *moreLikeThisToken
	_addQueryParameter func(interface{}) string
	_onDispose         func()
}

func NewMoreLikeThisScope(token *moreLikeThisToken, addQueryParameter func(interface{}) string, onDispose func()) *MoreLikeThisScope {
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

func (s *MoreLikeThisScope) WithOptions(options *MoreLikeThisOptions) {
	if options == nil {
		return
	}

	// force using *non* entity serializer here:
	optionsAsJson := ValueToTree(options)
	s._token.optionsParameterName = s._addQueryParameter(optionsAsJson)
}

func (s *MoreLikeThisScope) withDocument(document string) {
	s._token.documentParameterName = s._addQueryParameter(document)
}
