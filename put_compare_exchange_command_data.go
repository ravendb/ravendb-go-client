package ravendb

var _ ICommandData = &PutCompareExchangeCommandData{}

type PutCompareExchangeCommandData struct {
	CommandData

	index    int64
	document map[string]interface{}
}

func NewPutCompareExchangeCommandData(key string, value map[string]interface{}, index int64) (*PutCompareExchangeCommandData, error) {
	res := &PutCompareExchangeCommandData{
		CommandData: CommandData{
			ID:   key,
			Type: CommandCompareExchangePut,
		},
		index:    index,
		document: value,
	}
	return res, nil
}

func (d *PutCompareExchangeCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	js := d.baseJSON()
	js["Index"] = d.index
	js["Document"] = d.document
	return js, nil
}
