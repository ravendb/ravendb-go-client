package ravendb

type GroupByKeyToken struct {
	*QueryToken

	_fieldName     string
	_projectedName string
}

func NewGroupByKeyToken(fieldName string, projectedName string) *GroupByKeyToken {
	return &GroupByKeyToken{
		QueryToken: NewQueryToken(),

		_fieldName:     fieldName,
		_projectedName: projectedName,
	}
}

func GroupByKeyToken_create(fieldName string, projectedName string) *GroupByKeyToken {
	return NewGroupByKeyToken(fieldName, projectedName)
}

func (t *GroupByKeyToken) writeTo(writer *StringBuilder) {
	t.writeField(writer, firstNonEmptyString(t._fieldName, "key()"))

	if t._projectedName == "" || t._projectedName == t._fieldName {
		return
	}

	writer.append(" as ")
	writer.append(t._projectedName)
}
