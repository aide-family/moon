package timer

import "time"

// Matcher is an interface for objects that can match a specific time.
// It is primarily used to determine if a given time meets certain conditions.
type Matcher interface {
	// Match checks if the given time meets the conditions.
	// It returns true if the time matches, otherwise false.
	Match(time.Time) bool
}
