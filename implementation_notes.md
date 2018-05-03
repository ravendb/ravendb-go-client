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
```
``

In Java, the pattern is:
```
```

Result is statically typed because we can encode type of the result via generic
parametrization.

In Go we invert the logic: a command uses executor to do HTTP
and we have
