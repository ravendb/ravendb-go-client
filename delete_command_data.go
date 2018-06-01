package ravendb

type DeleteCommandData struct {
	*CommandData
}

// NewDeleteCommandData creates CommandData for Delete command
func NewDeleteCommandData(id string, changeVector string) ICommandData {
	res := &DeleteCommandData{
		&CommandData{
			Type:         CommandType_DELETE,
			ID:           id,
			ChangeVector: changeVector,
		},
	}
	return res
}

func (d *DeleteCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	return d.baseJSON(), nil
}
