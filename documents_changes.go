package ravendb

// ChangeType describes a type of a change in a document
type ChangeType int

const (
	// TODO: make those into a string? q q
	DocumentChangeDocumentDeleted ChangeType = iota
	DocumentChangeDocumentAdded
	DocumentChangeFieldChanged
	DocumentChangeNewField
	DocumentChangeRemovedField
	DocumentChangeArrayValueChanged
	DocumentChangeArrayValueAdded
	DocumentChangeArrayValueRemoved
	DocumentChangeFieldTypeChanged
	DocumentChangeEntityTypeChanged
)

// DocumentsChanges describes a change in a document
type DocumentsChanges struct {
	FieldOldValue interface{}
	FieldNewValue interface{}
	FieldOldType  JsonNodeType
	FieldNewType  JsonNodeType
	Change        ChangeType
	FieldName     string
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
