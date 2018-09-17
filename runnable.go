package ravendb

// Note: in Java Runnable is a class with run() function. In Go we use
// a void function. Instead of foo.run() we do foo()
type Runnable func()
