package ravendb

var _ ICommandData = &DeleteCompareExchangeCommandData{}

type DeleteCompareExchangeCommandData struct {
	CommandData

	index int64
}

func NewDeleteCompareExchangeCommandData(key string, index int64) (*DeleteCompareExchangeCommandData, error) {
	res := &DeleteCompareExchangeCommandData{
		CommandData: CommandData{
			ID:   key,
			Type: CommandCompareExchangeDelete,
		},
		index: index,
	}
	return res, nil
}

func (d *DeleteCompareExchangeCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	js := d.baseJSON()
	js["Index"] = d.index
	return js, nil
}
