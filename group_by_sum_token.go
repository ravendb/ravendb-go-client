package ravendb

type GroupBySumToken struct {
	*QueryToken

	_projectedName string
	_fieldName     string
}

func NewGroupBySumToken(fieldName string, projectedName string) *GroupBySumToken {
	return &GroupBySumToken{
		QueryToken: NewQueryToken(),

		_fieldName:     fieldName,
		_projectedName: projectedName,
	}
}

func GroupBySumToken_create(fieldName string, projectedName string) *GroupBySumToken {
	return NewGroupBySumToken(fieldName, projectedName)
}

func (t *GroupBySumToken) writeTo(writer *StringBuilder) {
	writer.append("sum(")
	writer.append(t._fieldName)
	writer.append(")")

	if t._projectedName == "" {
		return
	}

	writer.append(" as ")
	writer.append(t._projectedName)
}
