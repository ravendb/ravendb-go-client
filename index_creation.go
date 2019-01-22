package ravendb

// IndexCreationCreateIndexes creates indexes in store
// TODO: better name
func IndexCreationCreateIndexes(indexes []*AbstractIndexCreationTask, store *DocumentStore, conventions *DocumentConventions) error {

	if conventions == nil {
		conventions = store.GetConventions()
	}

	indexesToAdd := indexCreationCreateIndexesToAdd(indexes, conventions)
	op := NewPutIndexesOperation(indexesToAdd...)
	err := store.Maintenance().Send(op)
	if err == nil {
		return nil
	}

	// For old servers that don't have the new endpoint for executing multiple indexes
	for _, index := range indexes {
		err = index.Execute(store, conventions, "")
		if err != nil {
			return err
		}
	}
	return nil
}

func indexCreationCreateIndexesToAdd(indexCreationTasks []*AbstractIndexCreationTask, conventions *DocumentConventions) []*IndexDefinition {
	var res []*IndexDefinition
	for _, x := range indexCreationTasks {
		x.Conventions = conventions
		definition := x.CreateIndexDefinition()
		definition.Name = x.GetIndexName()
		pri := x.Priority
		if pri == "" {
			pri = IndexPriorityNormal
		}
		definition.Priority = pri
		res = append(res, definition)
	}
	return res
}
