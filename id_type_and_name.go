package ravendb

type idTypeAndName struct {
	id   string
	typ  CommandType
	name string
}

func newIDTypeAndName(id string, typ CommandType, name string) idTypeAndName {
	return idTypeAndName{
		id:   id,
		typ:  typ,
		name: name,
	}
}
