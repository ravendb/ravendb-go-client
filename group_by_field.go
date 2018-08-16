package ravendb

type GroupByField struct {
	fieldName     string
	projectedName string
}

func NewGroupByField() *GroupByField {
	return &GroupByField{}
}

func NewGroupByFieldWithName(fieldName string) *GroupByField {
	return &GroupByField{
		fieldName: fieldName,
	}
}

func NewGroupByFieldWithNameAndProjectedName(fieldName string, projectedName string) *GroupByField {
	return &GroupByField{
		fieldName:     fieldName,
		projectedName: projectedName,
	}
}

func (f *GroupByField) GetFieldName() string {
	return f.fieldName
}

func (f *GroupByField) setFieldName(fieldName string) {
	f.fieldName = fieldName
}

func (f *GroupByField) getProjectedName() string {
	return f.projectedName
}

func (f *GroupByField) setProjectedName(projectedName string) {
	f.projectedName = projectedName
}
