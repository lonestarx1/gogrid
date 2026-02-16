// Package team implements GoGrid's team (chat room) orchestration pattern.
// Multiple agents run concurrently on the same input, communicate via a
// pub/sub message bus, share state through a shared memory pool, and reach
// decisions through pluggable consensus strategies.
package team
