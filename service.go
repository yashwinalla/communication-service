package main

// Service struct holds all variables common to all handlers.
// That is why members have to be safe for concurrent use and do not cause race conditions!
type Service struct {
}
