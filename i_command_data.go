package ravendb

type ICommandData interface {
	getId() string
	GetName() string
	GetChangeVector() *string
	getType() CommandType
	serialize(conventions *DocumentConventions) (interface{}, error)
}

// CommandData describes common data for commands
type CommandData struct {
	ID           string
	Name         string
	ChangeVector *string
	Type         CommandType
}

func (d *CommandData) getId() string {
	return d.ID
}

func (d *CommandData) GetName() string {
	return d.Name
}

func (d *CommandData) getType() string {
	return d.Type
}

func (d *CommandData) GetChangeVector() *string {
	return d.ChangeVector
}

func (d *CommandData) baseJSON() ObjectNode {
	res := ObjectNode{
		"Id":           d.ID,
		"Type":         d.Type,
		"ChangeVector": d.ChangeVector,
	}
	return res
}
