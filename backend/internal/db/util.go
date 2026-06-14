package db

// nullStr converts an empty string to nil so it is stored as SQL NULL.
func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
