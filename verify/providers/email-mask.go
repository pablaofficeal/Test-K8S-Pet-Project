package providers

import "strings"

func MaskEmail(email string) string {
	at := strings.LastIndex(email, "@")
	if at < 1 {
		return "***"
	}
	return email[:1] + "***" + email[at:]
}