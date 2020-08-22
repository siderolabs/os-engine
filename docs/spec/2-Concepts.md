# Concepts

## Event

An event is a data structure that represents an event.

## Source

A source is an event generator.

## Sink

A sink is a consumer of events.

## Loader

A loader is anything that can load a source, or sink into the engine.

### Domain

A domain is a classification of an event based on factors such as

- The Linux subsytem from which the event is generated

## Capabilities

The capabilities of a machine are defined by:

- Hardware
- Kernel configuration <!-- Can we look at how the kernel generates /proc/config.gz and do something similar at boot time to parse out what sources to enable? -->
- SMBIOS

## Runtime

The runtime is a sink with the added responsibility of implementing an API for working with the state of the OS.
