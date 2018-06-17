package ravendb

type GroupByToken struct {
	*QueryToken

	_fieldName string
	_method    GroupByMethod
}

func NewGroupByToken(fieldName string, method GroupByMethod) *GroupByToken {
	return &GroupByToken{
		QueryToken: NewQueryToken(),

		_fieldName: fieldName,
		_method:    method,
	}
}

func GroupByToken_create(fieldName string) *GroupByToken {
	return GroupByToken_createWithMethod(fieldName, GroupByMethod_NONE)
}

func GroupByToken_createWithMethod(fieldName string, method GroupByMethod) *GroupByToken {
	return NewGroupByToken(fieldName, method)
}

func (t *GroupByToken) writeTo(writer *StringBuilder) {
	_method := t._method
	if _method != GroupByMethod_NONE {
		writer.append("Array(")
	}
	t.writeField(writer, t._fieldName)
	if _method != GroupByMethod_NONE {
		writer.append(")")
	}
}
