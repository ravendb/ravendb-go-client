package ravendb

type ChangeType int

const (
	DOCUMENT_DELETED ChangeType = iota
	DOCUMENT_ADDED
	FIELD_CHANGED
	NEW_FIELD
	REMOVED_FIELD
	ARRAY_VALUE_CHANGED
	ARRAY_VALUE_ADDED
	ARRAY_VALUE_REMOVED
	FIELD_TYPE_CHANGED
	ENTITY_TYPE_CHANGED
)

type DocumentsChanges struct {
	fieldOldValue interface{}

	fieldNewValue interface{}

	fieldOldType JsonNodeType

	fieldNewType JsonNodeType

	change ChangeType

	fieldName string
}

func (c *DocumentsChanges) getFieldOldValue() interface{} {
	return c.fieldOldValue
}

func (c *DocumentsChanges) setFieldOldValue(fieldOldValue interface{}) {
	c.fieldOldValue = fieldOldValue
}

func (c *DocumentsChanges) getFieldNewValue() interface{} {
	return c.fieldNewValue
}

func (c *DocumentsChanges) setFieldNewValue(fieldNewValue interface{}) {
	c.fieldNewValue = fieldNewValue
}

func (c *DocumentsChanges) getFieldOldType() JsonNodeType {
	return c.fieldOldType
}

func (c *DocumentsChanges) setFieldOldType(fieldOldType JsonNodeType) {
	c.fieldOldType = fieldOldType
}

func (c *DocumentsChanges) getFieldNewType() JsonNodeType {
	return c.fieldNewType
}

func (c *DocumentsChanges) setFieldNewType(fieldNewType JsonNodeType) {
	c.fieldNewType = fieldNewType
}

func (c *DocumentsChanges) getChange() ChangeType {
	return c.change
}

func (c *DocumentsChanges) setChange(change ChangeType) {
	c.change = change
}

func (c *DocumentsChanges) getFieldName() string {
	return c.fieldName
}

func (c *DocumentsChanges) setFieldName(fieldName string) {
	c.fieldName = fieldName
}

/*
   @Override
   public boolean equals(Object o) {
       if (this == o) return true;
       if (o == null || getClass() != o.getClass()) return false;

       DocumentsChanges that = (DocumentsChanges) o;

       if (fieldOldValue != null ? !fieldOldValue.equals(that.fieldOldValue) : that.fieldOldValue != null)
           return false;
       if (fieldNewValue != null ? !fieldNewValue.equals(that.fieldNewValue) : that.fieldNewValue != null)
           return false;
       if (fieldOldType != that.fieldOldType) return false;
       if (fieldNewType != that.fieldNewType) return false;
       if (change != that.change) return false;
       return fieldName != null ? fieldName.equals(that.fieldName) : that.fieldName == null;
   }
*/
