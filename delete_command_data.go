package ravendb

// DeleteCommandData represents data for delete command
type DeleteCommandData struct {
	CommandData
}

// NewDeleteCommandData creates ICommandData for Delete command
func NewDeleteCommandData(id string, changeVector string) ICommandData {
	var changeVectorPtr *string
	if changeVector != "" {
		changeVectorPtr = &changeVector
	}
	res := &DeleteCommandData{
		CommandData{
			Type:         CommandDelete,
			ID:           id,
			ChangeVector: changeVectorPtr,
		},
	}
	return res
}

func (d *DeleteCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	return d.baseJSON(), nil
}
