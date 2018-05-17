package ravendb

type IdTypeAndName struct {
	id   String
	typ  CommandType
	name String
}

func NewIdTypeAndName(id String, typ CommandType, name String) IdTypeAndName {
	return IdTypeAndName{
		id:   id,
		typ:  typ,
		name: name,
	}
}

// TODO: remove setter and getter functions after porting most of the code
func (t *IdTypeAndName) getId() String {
	return t.id
}

func (t *IdTypeAndName) setId(id String) {
	t.id = id
}

func (t *IdTypeAndName) getType() CommandType {
	return t.typ
}

func (t *IdTypeAndName) setType(typ CommandType) {
	t.typ = typ
}

func (t *IdTypeAndName) getName() String {
	return t.name
}

func (t *IdTypeAndName) setName(name String) {
	t.name = name
}
