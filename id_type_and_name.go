package ravendb

type IdTypeAndName struct {
	id   string
	typ  CommandType
	name string
}

func NewIdTypeAndName(id string, typ CommandType, name string) IdTypeAndName {
	return IdTypeAndName{
		id:   id,
		typ:  typ,
		name: name,
	}
}

// TODO: use NewIdTypeAndName instead
func IdTypeAndName_create(id string, typ CommandType, name string) IdTypeAndName {
	return NewIdTypeAndName(id, typ, name)
}
