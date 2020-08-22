// Package source defines the event sink, and offers the following default
// implementations:
//   - Debug: Prints human readable strings to standard output.
//   - gRPC: Sends events using remote procedure calls.
//   - Go plugin: Sends events to a go plugin.
package sink
