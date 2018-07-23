package ravendb

type GroupBy struct {
	field  string
	method GroupByMethod
}

func NewGroupBy() *GroupBy {
	// empty
	return &GroupBy{}
}

func (g *GroupBy) getField() string {
	return g.field
}

func (g *GroupBy) getMethod() GroupByMethod {
	return g.method
}

func GroupBy_field(fieldName string) *GroupBy {
	return &GroupBy{
		field:  fieldName,
		method: GroupByMethod_NONE,
	}
}

func GroupBy_array(fieldName string) *GroupBy {
	return &GroupBy{
		field:  fieldName,
		method: GroupByMethod_ARRAY,
	}
}
