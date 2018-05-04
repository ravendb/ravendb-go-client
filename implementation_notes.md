Go is a statically typed language without generics.

That means that code patterns that work well in a dynamicaly typed language (Python,
Ruby) or a statically typed language with generics (Java, C#) are akward or
impossible when transliterated to Go.

Go library follows structure and terminology of Python and Java libraries but
sometimes it must diverge.

To make future maintance easier, this documents implementation choices and why
they were made.

## Comands and RequestExecutor

A RavenCommand encapsulates a unique request/response interaction with
the server over HTTP. RavenCommand constructs HTTP request from command-specific
arguments and parses JSON response into a command-specific return value.

In Python, the pattern is:

```python
    database_names = store.maintenance.server.send(GetDatabaseNamesOperation(0, 3))
``

As everything in Python, the result is dynamically typed, but the caller knows
what to expect.

In Java, the pattern is:

```java
    GetDatabaseNamesOperation databaseNamesOperation = new GetDatabaseNamesOperation(0, 20);
    RavenCommand<String[]> command = databaseNamesOperation.getCommand(conventions);
    String[] databaseNames = executor.execute(command);
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
