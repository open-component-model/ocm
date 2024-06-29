// Package pubsub contains the
// handling required to connect OCM repositories to publish/subscribe
// infrastructures.
// A pubsub system is described by a dedicated specification (de-)serializable
// by a PubSubType object. The deserialized specification must implement the
// PubSubSpec interface. It has to be able to provide a pubsub adapter by
// providing an object implementing the PubSubMethod interface instantiated
// for a dedicated repository object. This object is then used by the
// OCM library (if provided) to generate appropriate events when adding/updating
// a component version for a repository.
//
// The known pubsub types can be registered for an OCM context. This registration
// mechanism uses a dedicated context attribute.
// The default type registry can be filled by init functions using the function
// RegisterType.
//
// The library allows to configure a pub/sub specification at repository level.
// Therefore, dedicated providers (interface Provider) are used, which are able
// to extract/provide a pubsub specification from/for a dedicated repository.
// The task of the provider is to handle the persistence of the serialized data
// of the specification at repository level. The provider just provides the
// specification data, it does not know anything about the types and implementations.7
//
// Providers are registered at an OCM context for a dedicated type of repository.
// (the spec kind of the repository spec used to describe the repository).
// The default provider registry can be filled by init functions using the function
// RegisterProvider.
//
// To configure dedicated contexts the attribute provided by For(ctx) can be modified
// contained registry objects.
package pubsub
