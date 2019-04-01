package ravendb

// Note: JavaScriptIndexCreationTask is Java's AbstractJavaScriptIndexCreationTask

type JavaScriptIndexCreationTask struct {
	*IndexCreationTask
	// TOD: maybe hide and add methods, like in Java
	Definition *IndexDefinition
}

func NewJavaScriptIndexCreationTask(indexName string) *JavaScriptIndexCreationTask {
	d := &IndexDefinition{
		LockMode: IndexLockModeUnlock,
		Priority: IndexPriorityNormal,
	}
	return &JavaScriptIndexCreationTask{
		IndexCreationTask: NewIndexCreationTask(indexName),
		Definition:        d,
	}
}

func (t *JavaScriptIndexCreationTask) IsMapReduce() bool {
	return t.Definition.Reduce != nil
}

func (t *JavaScriptIndexCreationTask) CreateIndexDefinition() *IndexDefinition {
	if t.IsMapReduce() {
		t.Definition.IndexType = IndexTypeJavaScriptMapReduce
	} else {
		t.Definition.IndexType = IndexTypeJavaScriptMap
	}
	if t.AdditionalSources != nil {
		t.Definition.AdditionalSources = t.AdditionalSources
	} else {
		t.Definition.AdditionalSources = map[string]string{}
	}

	return t.Definition
}
