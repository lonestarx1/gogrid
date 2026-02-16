// Package file provides a file-backed implementation of memory.Memory.
// Each key is stored as a separate JSON file with a metadata sidecar,
// making entries easy to inspect and avoiding file-level contention.
package file
