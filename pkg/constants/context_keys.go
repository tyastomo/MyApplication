package constants

// ContextKey type to be used for context values
type ContextKey string

const (
	// UserIDKey is the key for storing UserID in context locals
	UserIDKey ContextKey = "userID"
	// UserTypeKey is the key for storing UserType in context locals
	UserTypeKey ContextKey = "userType"
	// RequestIDKey is the key for storing RequestID in context locals (useful for logging/auditing)
	RequestIDKey ContextKey = "requestID"
)

// String returns the string representation of the ContextKey
func (c ContextKey) String() string {
	return string(c)
}
