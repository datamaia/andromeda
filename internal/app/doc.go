// Package app is layer L4 application composition: it wires the L2 infrastructure and L1
// contracts into runnable flows for the cmd/andromeda driver. It holds no domain logic of its
// own — it composes engines and adapters. At this stage it provides the environment
// diagnostic that exercises the MS-1 foundation end to end.
package app
