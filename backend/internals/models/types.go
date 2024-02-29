package models

func RemoveStringFromStringSlice(slice []string, target string) []string {
	var result []string

	for _, s := range slice {
		if target != s {
			result = append(result, s)
		}
	}

	return result
}
