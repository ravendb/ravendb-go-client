package ravendb

var _ IObserver = &ActionBasedObserver{}

// Note: use NewActionBasedObserver instead Observers.create()

type ActionBasedObserver struct {
	action func(interface{})
}

func NewActionBasedObserver(action func(interface{})) *ActionBasedObserver {
	return &ActionBasedObserver{
		action: action,
	}
}

func (o *ActionBasedObserver) OnNext(value interface{}) {
	o.action(value)
}

func (o *ActionBasedObserver) OnError(err error) {
	//empty
}

func (o *ActionBasedObserver) OnCompleted() {
	//empty
}
