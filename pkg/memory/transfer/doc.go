// Package transfer provides state ownership management for GoGrid's
// pipeline pattern. TransferableState wraps a memory.Memory with a
// generation counter so that when state is transferred from one agent
// to the next, the previous owner's Handle becomes invalid. This prevents
// confused deputy problems where a stale agent modifies state it no
// longer owns.
package transfer
