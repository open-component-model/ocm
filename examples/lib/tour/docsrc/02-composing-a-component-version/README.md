{{compose-compvers}}
# Composing a Component Version

This tour illustrates the basic usage of the API to
create/compose component versions.

It covers two basic scenarios:
- [`basic`](01-basic-componentversion-creation.go) Create a component version stored in the filesystem
- [`compose`](02-composition-version.go) Create a component version stored in memory using a non-persistent composition version.

## Running the example

You can just call the main program with the scenario as argument. Configuration is not required.

## Walkthrought



### Basic Component Version Creation
The first variant just creates a new component version
in an OCM repository. To avoid the requirement for 
credentials a filesystem based repository is created, using
the *Common Transport Format* (CTF).

As usual, we start with getting access to an OCM context
object

```go
{{include}{../../02-composing-a-component-version/01-basic-componentversion-creation.go}{default context}}
```

To compose and store a new component version
we finally need some OCM repository to
store the component, The most simple
external repository could be the filesystem.
OCM defines a distribution format, the
Common Transport Format (CTF) for this,
which is an extension of the OCI distribution
specification.
There are three flavours, *Directory*, *Tar* or *TGZ*.
The implementation provides a regular OCM repository
interface, like used in the previous example.

```go
{{include}{../../02-composing-a-component-version/01-basic-componentversion-creation.go}{create ctf}}
```

Once we have a repository we can compose a new version.
First we create a new version backed by this repository.
The result is a memory based representation not yet persisted.

```go
{{include}{../../02-composing-a-component-version/01-basic-componentversion-creation.go}{new version}}
```


### Composition Environment