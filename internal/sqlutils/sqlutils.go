package sqlutils

func GetPlaceholders(n int) []string {
	placeholders := make([]string, n)
	for i := range placeholders {
		placeholders[i] = "?"
	}

	return placeholders
}
