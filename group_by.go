package ravendb

// GroupBy represents arguments to "group by" query
type GroupBy struct {
	Field  string
	Method GroupByMethod
}

// NewGroupByField returns new GroupBy for a field
func NewGroupByField(fieldName string) *GroupBy {
	return &GroupBy{
		Field:  fieldName,
		Method: GroupByMethodNone,
	}
}

// NewGroupByField returns new GroupBy for an array
func NewGroupByArray(fieldName string) *GroupBy {
	return &GroupBy{
		Field:  fieldName,
		Method: GroupByMethodArray,
	}
}
