// +build for_tests

package ravendb

// this file exposes functionality that is only meant to be used
// in tests. This code is only compiled when "-tags for_tests"
// option is used

func (c *DocumentConventions) GetCollectionName(entityOrType interface{}) string {
	return c.getCollectionName(entityOrType)
}

func (s *DocumentSession) ForTestsSaveChangesGetCommands() ([]ICommandData, error) {
	result, err := s.prepareForSaveChanges()
	if err != nil {
		return nil, err
	}
	res := append(result.sessionCommands, result.deferredCommands...)
	return res, nil
}
