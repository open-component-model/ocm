// Package transferhandler provides the API for transfer handlers used
// during the transfer process of an OCM component.
// There is a common generic transfer functionality in package transfer,
// which can be used to transfer component versions from one OCM repository
// to another one. To flexible enough for various use cases this generic handling
// uses extension points to control the concrete transfer behaviour.
// These extension points can be implemented by dedicated transfer handlers
// to influence the transfer process.
//
// For example:
//
//   - In a dedicated scenario only components from a dedicated provider should
//     not be transferred per value in a transfer operation.
//   - Another scenario could require to replace dedicated access methods.
//
// Such specific logic is hard to formalize, but nevertheless the basic transfer
// flow is always the same. Therefore, it makes no sense to reimplement the flow
// just because some inner process steps or decisions should look slightly
// different. Because of this transfer handlers are introduces, which allow to
// separate such decisions and step implementations form the generic flow and
// to provided scenario specific implementations with reimplementing the complete
// transitive transfer handling.
//
// For common use cases two standard handlers are provided, which can be used
// to formally describe typical transfer scenarios:
//   - package [github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard]
//     provides a standard behaviour configurable
//     by some common options, like transport-by-value.
//   - package [github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/spiff]
//     provides a handler configurable by a [spiff] script.
//
// Transfer handlers may accept transfer options to be configured.
// For operations not taking a transferhandler but only transfer options,
// the given options are used to find the best matching transfer handling accepting
// those options. Therefore, standard transfer handles can be registered.
// Besides this any user defined transfer handler may be used by using operation
// flavors directly accepting a transfer handler.
//
// [spiff]: https://github.com/mandelsoft/spiff/README.md
package transferhandler
