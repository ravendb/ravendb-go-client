This document tries to explain how we have to handle maps in Go code because it's a bit special.

See also https://github.com/ravendb/ravendb-go-client/issues/105 

Maps have to be passed by a pointer to both `Store` and `Load` methods.

This is different from structs where `Store` takes a `*Foo` and `Load` takes `**Foo`.

In Go a map value, under the hood, is a pointer.
 
However, we don't have access to that pointer (without using hacks that depend on internals of the runtime/compiler, which might change in the future).

One result of that is that Go only allows to compare map value to a nil but not to another map.

The only way to have a stable reference to a map is to use a pointer to it.

This is similar to taking a pointer to a struct but it does look weird.



