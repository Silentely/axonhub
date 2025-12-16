package xptr

// IntPtr returns a pointer to the given int value.
// Useful for creating pointers to int literals in test cases.
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to the given bool value.
func BoolPtr(b bool) *bool {
	return &b
}

// StringPtr returns a pointer to the given string value.
func StringPtr(s string) *string {
	return &s
}
