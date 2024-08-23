# Type Handlers

The task of type handlers is to resolve element specifications typically
taken from the command line arguments to a sequence of objects that are
the processed to processing steps in a processing chain.

As such they work as data source for processing chains.

The are controlled by a generic function family *HandleOutput(s)* that handle
the API calls for the type handlers according to the actual set of
element specifications.

There are two basic methods:

- All() - Without an element spec all possible elements are returned
- Get() - If specs are given for every spec the Get method is called
  to resolve the spec to a list of elements.
