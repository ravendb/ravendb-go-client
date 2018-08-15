package ravendb

type ChangeType int

const (
	DocumentsChanges_ChangeType_DOCUMENT_DELETED ChangeType = iota
	DocumentsChanges_ChangeType_DOCUMENT_ADDED
	DocumentsChanges_ChangeType_FIELD_CHANGED
	DocumentsChanges_ChangeType_NEW_FIELD
	DocumentsChanges_ChangeType_REMOVED_FIELD
	DocumentsChanges_ChangeType_ARRAY_VALUE_CHANGED
	DocumentsChanges_ChangeType_ARRAY_VALUE_ADDED
	DocumentsChanges_ChangeType_ARRAY_VALUE_REMOVED
	DocumentsChanges_ChangeType_FIELD_TYPE_CHANGED
	DocumentsChanges_ChangeType_ENTITY_TYPE_CHANGED
)

type DocumentsChanges struct {
	fieldOldValue interface{}

	fieldNewValue interface{}

	fieldOldType JsonNodeType

	fieldNewType JsonNodeType

	change ChangeType

	fieldName string
}

func NewDocumentsChanges() *DocumentsChanges {
	return &DocumentsChanges{}
}

func (c *DocumentsChanges) GetFieldOldValue() interface{} {
	return c.fieldOldValue
}

func (c *DocumentsChanges) setFieldOldValue(fieldOldValue interface{}) {
	c.fieldOldValue = fieldOldValue
}

func (c *DocumentsChanges) GetFieldNewValue() interface{} {
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

func (c *DocumentsChanges) GetChange() ChangeType {
	return c.change
}

func (c *DocumentsChanges) setChange(change ChangeType) {
	c.change = change
}

func (c *DocumentsChanges) GetFieldName() string {
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
