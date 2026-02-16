// Package shared provides a thread-safe shared memory pool for GoGrid's
// team/chat-room agent patterns. Multiple agents can read and write to the
// same memory store concurrently, with optional change notifications via
// Go channels. NamespacedView provides transparent key prefixing so agents
// can share a pool without key collisions.
package shared
