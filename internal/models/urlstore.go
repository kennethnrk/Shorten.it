package models

import (
	"github.com/gocql/gocql"

	"shortenit/internal/utils"
)

func FetchLongURL(session *gocql.Session, code string) (string, error) {
	var longURL string
	err := session.Query("SELECT long_url FROM long_urls_by_code WHERE code = ?", code).Scan(&longURL)
	if err == gocql.ErrNotFound {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return longURL, nil
}

func FetchShortURL(session *gocql.Session, longURL string) (string, error) {

	longURLHash := utils.GenerateLongURLHash(longURL)
	var code string
	err := session.Query("SELECT code FROM short_urls_by_long_url WHERE long_url_hash = ?", longURLHash).Scan(&code)
	if err == gocql.ErrNotFound {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return code, nil
}

func InsertNewURL(session *gocql.Session, longURL string, shortURL string) error {

	longURLHash := utils.GenerateLongURLHash(longURL)
	err := session.Query("INSERT INTO short_urls_by_long_url (long_url_hash, code, created_at) VALUES (?, ?,  dateof(now()))", longURLHash, shortURL).Exec()
	if err != nil {
		return err
	}

	err = session.Query("INSERT INTO long_urls_by_code (code, long_url, created_at) VALUES (?, ?, dateof(now()))", shortURL, longURL).Exec()
	if err != nil {
		return err
	}
	return nil
}
