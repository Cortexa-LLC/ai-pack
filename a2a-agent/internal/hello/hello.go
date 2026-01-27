// Package hello provides simple greeting functionality.
package hello

import "strings"

// HelloWorld returns a greeting message.
// If name is empty, it returns a default "Hello, World!" greeting.
// If name is provided, it returns a personalized greeting with the name trimmed of whitespace.
func HelloWorld(name string) string {
	// Trim whitespace from the input
	name = strings.TrimSpace(name)

	// Return default greeting if name is empty
	if name == "" {
		return "Hello, World!"
	}

	// Return personalized greeting
	return "Hello, " + name + "!"
}
