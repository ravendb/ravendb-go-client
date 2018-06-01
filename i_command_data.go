package ravendb

type ICommandData interface {
	getId() String
	getName() String
	getChangeVector() String
	getType() CommandType
	serialize(conventions *DocumentConventions) (interface{}, error)
}

// CommandData describes common data for commands
type CommandData struct {
	ID           string
	Name         string
	ChangeVector string
	Type         CommandType
}

func (d *CommandData) getId() string {
	return d.ID
}

func (d *CommandData) getName() string {
	return d.Name
}

func (d *CommandData) getType() string {
	return d.Type
}

func (d *CommandData) getChangeVector() string {
	return d.ChangeVector
}

func (d *CommandData) baseJSON() ObjectNode {
	res := ObjectNode{
		"Id":   d.ID,
		"Type": d.Type,
	}
	// TODO: ChangeVector is more subtle (null vs "" vs != "")
	if d.ChangeVector != "" {
		res["ChangeVector"] = d.ChangeVector
	}
	return res
}
