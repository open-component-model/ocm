# Resource Management

{{resmgmt}}

This tour illustrates the basic contract to
correctly work with closeable object references used
in the library.

Many objects provided by the library offer some kind of resource management. In the [first example]({{getting-started}}), this is an
OCM repository, the OCM component, component version and the access method.
Another important kind of objects are the `BlobAccess` implementations.

Those objects may use external resources, like temporary file system content or caches. To get rid of those resources again, they offer a `Close` method.

To achieve the possibility to pass those objects around in non-functional call contexts they feature some kind of resource management. It allows to handle
the life cycle of the resource in a completely local manner. To do so, a second method `Dup` is offered, which provides an independent reference to the original resources, which can be closed separately.
The possible externally held resource are released with the close of the last reference.

This offers a simple contract to handle resources in functions or object methods:

1. a function creating such an object is responsible for the life cycle of its reference

    - if the object is returned, this responsibility is passed to its caller

      ```go
      func f() (Object, error) {
          o, err:= Create()
          if err != nil {
              return nil, err
          }
          o.DoSomeThing()
          DoSomeThingOther(o)
          return o, nil
      }
      ```

    - otherwise, it must be closed at the end of the function (or if it is not used anymore)

      ```go
      func f() error {
          o, err:= Create()
          if err != nil {
              return err
          }
          defer o.Close()
          o.DoSomeThing()
          DoSomeThingOther(o)
      }
      ```

   The object may be passed to any called function without bothering what this function does with this reference.

2. a function receiving such an object from a function as result it inherits   the responsibility to close it again (see case 1)

3. a function receiving such an object as an argument can freely use  it and a pass it around.

    ```go
    func f(o Object) {
        o.DoSomeThing()
        DoSomeThingOther(o)
    }
    ```

   If it decides to store the reference in some state, it must use an own reference for this, obtained by a call to `Dup`. After obtaining an own reference the used storage context is responsible to close it again. It should never close the obtained reference, because the caller is responsible for this.

    ```go
    func (r *State) f(o Object) (err error) {
        r.obj, err = o.Dup()
        return err
    }
   
    func (r *State) Close() error {
        if r.obj == nil {
            return nil
        }
        return r.obj.Close()
    }
    ```

## Running the example

You can call the main program without any argument.

## Walkthrough

The example is based on the initial [getting started scenario]({{getting-started}}).
It separates the resource gathering from the handling of the found resources.

```go
{{include}{../../07-resource-management/example.go}{decouple}}
```

The resources are provided by an array of the interface `Resource`:

```go
{{include}{../../07-resource-management/example.go}{resource interface}}
```

It encapsulates the technical resource handling
and offers a `Close` method, also, to release potential local resources.

The example provides one implementation, using the original access method
to cache the data to avoid additional copies.

```go
{{include}{../../07-resource-management/example.go}{resource implementation}}
```

The `AddDataFromMethod` uses `Dup` to provide an own reference to the
access method, which is stored in the provided resource object.
It implements the `Close` method to release this cached content, again.
The responsibility for this reference is taken by the `resource`object.

In the `GatherResources` function, a repository access is created.
It is not forwarded, and therefore closed, again, in this function.

```go
{{include}{../../07-resource-management/example.go}{repository}}
```

The same is done for the component version lookup.

```go
{{include}{../../07-resource-management/example.go}{lookup component}}
```

Then the resource `factory` is used to create the `Resource` objects for
the resources found in the component version.

```go
{{include}{../../07-resource-management/example.go}{resources}}
```

Because the function cannot know what happens behind the call to
`AddDataFromMethod`, it just closes everything what is created
in the function, this also includes the access method (`m`).

Finally, it returns the resource array after all locally created
references are correctly closed.
The provided `Resource` objects have taken the responsibility for
keeping their own references.

The resource handling function just uses the resources.

```go
{{include}{../../07-resource-management/example.go}{handle}}
```

The responsibility for closing the resources has been passed to
the `ResourceManagement` functions, which calls the gather and
the handling function. Therefore, it calls the `Resource.Close`
function before finishing.

The final output of this example looks like:

```yaml
{{execute}{go}{run}{../../07-resource-management}}
```
