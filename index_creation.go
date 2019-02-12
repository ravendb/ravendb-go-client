package ravendb

func indexCreationCreateIndexesToAdd(indexCreationTasks []*IndexCreationTask, conventions *DocumentConventions) []*IndexDefinition {
	var res []*IndexDefinition
	for _, x := range indexCreationTasks {
		x.Conventions = conventions
		definition := x.CreateIndexDefinition()
		definition.Name = x.IndexName
		pri := x.Priority
		if pri == "" {
			pri = IndexPriorityNormal
		}
		definition.Priority = pri
		res = append(res, definition)
	}
	return res
}
