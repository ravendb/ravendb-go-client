package ravendb

// result should be *<type> and we'll do equivalent of: *result = <type>
func makeStructFromJSONMap2(result interface{}, js map[string]interface{}) error {
	// TODO: not sure if should accept result of *map[string]interface{} or map[string]interface{}
	if res, ok := result.(*map[string]interface{}); ok {
		*res = js
		return nil
	}
	if res, ok := result.(map[string]interface{}); ok {
		for k, v := range js {
			res[k] = v
		}
		return nil
	}

	d, err := jsonMarshal(js)
	if err != nil {
		return err
	}
	return jsonUnmarshal(d, result)
}

func makeStructFromJSONMap3(result interface{}, js map[string]interface{}) error {
	return nil
}
