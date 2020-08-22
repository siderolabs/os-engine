# Goals

## Events

The scope of events is anything related to the operating system that can be expressed as an event.
These events will most commonly originate from an event within the kernel.
Some examples of this might be:

- When a block device is added, updated, or removed
- When a network interface is up, or down

## Observability

The scope of observability is any type of metric, tracing, or log related to the performance of the operating system.
The set of obeservability tooling are things deemed core to the observability of any implementation of the os-engine.

## Pluggable

In order to be flexible, the spec must define a pluggable system by which any implementation of os-engine can expose a way to plugin to the system.
