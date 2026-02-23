// Package auth contains authentication-related types and utilities,
// including JWT handling and role definitions.
package auth

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// IsAdmin reports whether the given role string corresponds to an administrator.
func IsAdmin(role string) bool {
	return role == RoleAdmin
}

// IsValid reports whether the given role string is a known/allowed value.
func IsValid(role string) bool {
	switch role {
	case RoleUser, RoleAdmin:
		return true
	default:
		return false
	}
}
