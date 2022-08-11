### Suport for TOI Executors

This package provides a generic command line tool support to provide
TOI executor CLIs.

Such a CLI wraps an executor specific executor function. If no options
are passed it complies to the TOI image binding contract.

For development purposes it can be called with a bunch of option to fake
the file system binding and redirect it to explicitly specified files.

It provides a contract to the executor function of type

<center>
<pre>
func(options *ExecutorOptions) error
</pre>
</center>

 which already reolves all those dependencies by providing an
 [ExecutorOptions](support.go#:~:text=type%20ExecutorOptions%20struct,%7D)
 object with the prepared contract data, including access to to
 the `ocm.ComponentVersionAccess` of the component version providing
 the package to work on.

 A typical `main` function of an executor could then look like:

 <pre>
     package main

     import (
        "os"

        "github.com/open-component-model/ocm/pkg/contexts/clictx"

        "yourpackage"
     )

     func main() {
        ctx:=clictx.New()
        // special configuration of the context. e.g. setting a virstual filesystem
        c := support.NewCLICommand(ctx.OCMContext(), "your executor name", yourpackage.ExecutorFunction)
        if err := c.Execute(); err != nil {
            os.Exit(1)
        }
     }
 </pre>


