package utils

func HasBackendChanged(oldList, newList []string) bool {
	if len(oldList) != len(newList) {
		return true
	}

	existing := make(map[string]bool, len(oldList))
	for _, b := range oldList {
		existing[b] = true
	}

	for _, b := range newList {
		if !existing[b] {
			return true
		}
	}

	return false
}
