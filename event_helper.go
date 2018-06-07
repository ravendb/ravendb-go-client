package ravendb

// TODO:
// public static <T extends EventArgs> void invoke(List<EventHandler<T>> delegates, Object sender, T event) {

func EventHelper_invoke(actions []Consumer, argument interface{}) {
	for _, action := range actions {
		action.accept(argument)
	}
}
