package utils

func ValidAccount(account string) bool {
	if len(account) < 3 || len(account) > 64 {
		return false
	}
	for _, char := range account {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') {
			return false
		}
	}
	return true
}
