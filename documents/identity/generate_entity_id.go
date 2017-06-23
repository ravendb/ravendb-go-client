package identity

import (
	"reflect"
	"strings"
)

//Attempts to get the document ID from an instance
func LookupIdFromInstance(entity interface{}) (id string, ok bool){
	if entity == nil{
		return nil, false
	}
	identityField, ok := getIdentityField(reflect.TypeOf(entity))
	if ok{
		entityElem := reflect.ValueOf(&entity).Elem()
		fieldVal := entityElem.FieldByIndex(identityField.Index)
		if fieldVal.CanInterface(){
			if fieldVal.Kind() != reflect.String {
				return nil, false
			}
			return string(fieldVal.Interface()), true
		}
	}
	id = reflect.TypeOf(entity).Name()
	if id == ""{
		return nil, false
	}

	return id, true
}

func getIdentityField(entType reflect.Type) (reflect.StructField, bool){
	propertyFieldIdx, ok := lookupIdentityPropertyIdxByTag(entType)
	if ok{
		return entType.Field(propertyFieldIdx), true
	}
	id, ok := entType.FieldByName("Id")
	if ok{
		return id, true
	}
	return nil, false
}

func lookupIdentityPropertyIdxByTag(entityType reflect.Type) (int, bool){
	for i := 0; i < entityType.NumField(); i++ {
		val := entityType.Field(i).Tag.Get("ravendb")
		if strings.HasSuffix(val, "id") || strings.Contains(val, "id,"){
			return i, true
		}
	}
	return nil, false
}