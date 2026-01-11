package utils

import (
	"regexp"
)

func ValidateLongURL(longURL string) bool {
	if longURL == "" {
		return false
	}
	re := regexp.MustCompile(`^https?:\/\/(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*)$`)
	if !re.MatchString(longURL) {
		return false
	}

	return true
}

func ValidateShortURL(shortURL string) bool {
	// short URL must be only base62 characters
	re := regexp.MustCompile(`^[0-9a-zA-Z]+$`)
	if !re.MatchString(shortURL) {
		return false
	}
	return true
}
