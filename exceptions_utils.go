package ravendb

func acceptError(action func() error) error {
	err := action()
	if err != nil {
		return unwrapError(err)
	}
	return nil
}

func unwrapError(e error) error {
	return e
	/*
		TODO: implement me
		if (e instanceof ExecutionException) {
			ExecutionException computationException = (ExecutionException) e;
			return unwrapError(computationException.getCause());
		}

		if (e instanceof RuntimeError) {
			return (RuntimeError)e;
		}

		return new RuntimeError(e);
	*/
}
