package lib

func IsValidColor(value string) bool {
	if len(value) != 3 && len(value) != 6 {
		return false
	}
	for _, c := range value {
		if (c < '0' || c > '9') && (c < 'A' || c > 'F') {
			return false
		}
	}

	return true
}
