Go is a statically typed language without generics.

That means that code patterns that work well in a dynamicaly typed language (Python,
Ruby) or a statically typed language with generics (Java, C#) are akward or
impossible when transliterated to Go.

Go library follows structure and terminology of Python and Java libraries but
sometimes it must diverge.

To make future maintance easier, this documents implementation choices and why
they were made.

## Java OOP vs. Go

Java has inheritance with ability to make some functions virtual in Base class and over-ride them in Derived classes.

Go only has embedding. A Derived struct can embed Base struct and will "inherit" fields and methods of the Base.

Go has interfaces which allows virtual functions. You can define an interface Foo, implement it by Bar1 and Bar2 structs. Function that takes Foo as an argument can receive Bar1 and Bar2 and will call the right virtual functions on them.

One might think that embedding + interface can be used to implement Java inheritence:
* define interface Foo
* have Base struct implement it
* embed Base in Derived struct
* over-write some interface (virutal) functions in Derived

There is a subtle but important difference.

if `Base.virt()` is a virtual function over-written by `Derived.virt()`, a function implemented on `Base` class will call `Derived.virt()` if the object is actually `Derived`.

It makes sense within the design. `Base` is embedded in `Derived`. `Derived` has access to `Base` but not the other way around. `Base` has no way to call code in `Derived`.

To put it differently:
* in Java, a virtual table is part of class and carried by Object itself. Virtual calls can therefore always be resolved
* in Go, a virtual table is carried as a separate interface type, which combines a value and its type information (including virtual table). We can only resolve virtual calls from interface type. Once virtual method is resolved it operates on a concrete type and only has access to that type

For example, in Java `RavenCommand.processResponse` calls virtual functions of derived classes. That couldn't be done in Go.

## Comands and RequestExecutor

A RavenCommand encapsulates a unique request/response interaction with
the server over HTTP. RavenCommand constructs HTTP request from command-specific
arguments and parses JSON response into a command-specific return value.

In Python, the pattern is:

```python
    database_names = store.maintenance.server.send(GetDatabaseNamesOperation(0, 3))
```

As everything in Python, the result is dynamically typed, but the caller knows
what to expect.

In Java, the pattern is:

```java
    GetDatabaseNamesOperation databaseNamesOperation = new GetDatabaseNamesOperation(0, 20);
    RavenCommand<string[]> command = databaseNamesOperation.getCommand(conventions);
    string[] databaseNames = executor.execute(command);
```

Result is statically typed because we can encode type of the result via generic
parametrization of RavenCommand.

Go has no generics so we can't have `RavenCommand` sublclasses specialized by return type.

We could mimic dynamic typing of Python and define interface for `RavenCommand`
which returns a parsed result for each command as `interface{}` but that would
be bad Go code. We want a staticly typed result.

So we invert the logic.

`RavenCommand` is just a struct that holds all information needed to construct
an HTTP request to the server.

For each command we have `NewFooCommand` (e.g. `NewGetClusterTopologyCommand`)
which creates `RavenCommand` from command-specific arguments.

For each command we also have `ExecuteFooCommand(executor, cmd)`
(e.g. `ExecuteGetClusterTopologyCommand`) which takes an abstract executor
that takes `RavenCommand`, does HTTP request and returns HTTP response.

`ExecuteFooCommadn` returns a command-specific result based on parsing JSON
response from the server.

The simplest implemention of executor runs the code against a single server.

Another implementation will adapt `RequestsExecutor` logic.

## How I port Java tests

Java runs tests in a fixed but unpredictable order. For easier debugging (e.g. when comparing recorded HTTP traffic) I want Go tests to run in the same order as Java tests.

I instrument Java test with System.out.println() to print name of executed test.

For e.g. `TrackEntityTest.java` I create `track_entity_test.go` (files that end with `_test.go` are only compiled when running tests).

`TracEntityTest` Java class is `TestTrackEntity()`. Each function in the form `Test*(*testing.T)` is a unique test for `go test`.

Each Java class method becomes a function e.g. `TrackEntityTest.deletingEntityThatIsNotTrackedShouldThrow` => `trackEntityTest_deletingEntityThatIsNotTrackedShouldThrow`.

Usually in Go each function would be a separate test function but to have control over test invocation order, they're part of `TestTrackEntity` test function.

To get HTTP logs for Java I add the test to `run_tests.go` to log to `trace_track_entity_java.txt` and call `./run_java_tests.sh`.

I port the tests and run them, also capturing HTTP logs to `trace_track_entity_go.txt`.

## How I debug tests

I use Visual Studio Code as an editor.

It's Go extension has a support for running individual tests (see https://www.notion.so/Debugging-tests-0f731a22d6154a7ba38a8503227b593d) so I set the desired breakpoints to step through the code and use that.

Other editors also support Go but I'm not familiar with them.

## Why no sub-packages?

Java code is split into multiple packages/sub-directories. Why not mimic that?

Go packages have restrictions: they can't have circular references.

Java code has lots of mutual-references between packages so it's impossible to
replicate its structure in Go.

## Enums

Go doesn't have enumes.

Java enums are represented as constants. Those that are `@UseSharpEnum` are typed as string. In other words, this:

```java
@UseSharpEnum
public enum FieldStorage {
    YES,
    NO
}
```

Is turned into this:
```go
type FieldStorage = string

const (
	FieldStorage_YES = "Yes"
	FieldStorage_NO  = "No"
)
```

## Statically ensuring a type implements an interface

Go implements duck-typing of interfaces i.e. a struct doesn't have to declare
that it implements an interface. That opens up a possibility of not
implementing an interface correctly.

A simple trick to ensure that a struct implements interface:

```go
var _ IVoidMaintenanceOperation = &PutClientConfigurationOperation{}
```

## toString()

Go has a `fmt.Stringer` interface with `String()` method but basic types (`int`, `float64` etc.) don't implement it (and we can't add methods to existing types).

Instead of `Object.toString` we can use `fmt.Sprintf("%#v", object)` which will use `String()` method if available and will format known types (including basic types) as their Go literal representation (most importantly it quotes strings so string `foo` has literal representation as `"foo"`).

To avoid quoting strings, use `%v` or `%s`.
