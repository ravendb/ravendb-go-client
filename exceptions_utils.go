package ravendb

func ExceptionsUtils_accept(action func() error) error {
	err := action()
	if err != nil {
		return ExceptionsUtils_unwrapException(err)
	}
	return nil
}

func ExceptionsUtils_unwrapException(e error) error {
	return e
	/*
		TODO: implement me
		if (e instanceof ExecutionException) {
			ExecutionException computationException = (ExecutionException) e;
			return unwrapException(computationException.getCause());
		}

		if (e instanceof RuntimeError) {
			return (RuntimeError)e;
		}

		return new RuntimeError(e);
	*/
}
