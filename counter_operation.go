package ravendb

type CounterOperation struct {
	typ          CounterOperationType
	counterName  string
	delta        int64
	changeVector string
}

func (o *CounterOperation) serialize(conventions *DocumentConventions) map[string]interface{} {
	res := map[string]interface{}{}
	res["Type"] = o.typ
	res["CounterName"] = o.counterName
	res["Delta"] = o.delta
	return res
}

/*
public class CounterOperation {
    public static CounterOperation create(String counterName, CounterOperationType type) {
        CounterOperation operation = new CounterOperation();
        operation.setCounterName(counterName);
        operation.setType(type);
        return operation;
    }

    public static CounterOperation create(String counterName, CounterOperationType type, long delta) {
        CounterOperation operation = new CounterOperation();
        operation.setCounterName(counterName);
        operation.setType(type);
        operation.setDelta(delta);
        return operation;
    }
}

*/
