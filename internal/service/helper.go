package service

// Helper function to safely dereference string pointers for logging
func safePtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
