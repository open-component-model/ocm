// Package plugin maps a Go plugin interface to a technical command line interface. Thereby, it allows to embed plugins
// and their provided functionality into the library. It is the basis for adapter implementations for accessmethods,
// downloaders, uploaders, blob handlers, actions, value mergers and label merge specifications.
// Those adapters can be found in the subpackage "plugin" of their dedicated functional area packages (e.g.
// pkg/contexts/ocm/accessmethods/plugin).
package plugin
