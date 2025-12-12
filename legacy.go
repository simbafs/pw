package main

import "log/slog"

// DEPRECATED: legacy mode is not implemented and returns a static password.
// This is a security vulnerability and should not be used.
// This function exists only for backward compatibility during migration.
func legacy(input string) string {
	slog.Warn("SECURITY WARNING: Legacy mode is deprecated and insecure - returns static password")
	return "this_is_a_dummy_legacy_password"
}
