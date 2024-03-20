package models

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	DOMAIN          = "url.shorten/"
	EXPIRY_DURATION = 60
)

type URLModel struct {
	DB *sql.DB
}

type URL struct {
	LongURL    string    `json:"long_url"`
	ShortURL   string    `json:"short_url"`
	ExpiryTime time.Time `json:"expiry_time"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time
}

func (uModel URLModel) findURL(url string) (URL, error) {
	urlRow := URL{}

	row := uModel.DB.QueryRow("SELECT long_url, short_url, expiry_time, created_at, updated_at from urls where long_url = ? LIMIT 1", url)

	err := row.Scan(&urlRow.LongURL, &urlRow.ShortURL, &urlRow.ExpiryTime, &urlRow.CreatedAt, &urlRow.UpdatedAt)
	if err != nil {
		log.Printf("Error Scanning database row at findURL \n %s", err)
		return urlRow, err
	}

	return urlRow, nil
}

func (uModel URLModel) Unshorten(shortURL string) (*URL, error) {
	urlRow := URL{}
	row := uModel.DB.QueryRow("SELECT long_url, short_url, expiry_time, created_at, updated_at from urls where short_url = ?", shortURL)

	err := row.Scan(&urlRow.LongURL, &urlRow.ShortURL, &urlRow.ExpiryTime, &urlRow.CreatedAt, &urlRow.UpdatedAt)
	if err != nil {
		log.Printf("Error Scanning database row at Unshorten \n %s", err)
		return nil, err
	}

	if isExpired(urlRow) {
		return nil, errors.New("Short URL already expired!")
	}

	return &urlRow, nil
}

func (uModel URLModel) Shorten(url string) (*URL, error) {
	hasher := md5.New()
	hasher.Write([]byte(url))
	hash := hex.EncodeToString(hasher.Sum(nil))
	urlRow, err := uModel.findOrCreate(url, hash)
	if err != nil {
		log.Printf("Error returned from findOrCreate \n %s", err)
		return nil, err
	}
	return urlRow, nil
}

func isExpired(urlRow URL) bool {
	if urlRow.ExpiryTime.Before(time.Now()) {
		return true
	}
	return false
}

func (uModel URLModel) addRecord(longURL, hash string) (*URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	shortURL := DOMAIN + hash
	expiryTime := time.Now().Add(time.Minute * EXPIRY_DURATION)
	res, err := uModel.DB.ExecContext(ctx, "INSERT INTO urls (long_url, short_url, expiry_time, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		longURL, shortURL, expiryTime, time.Now(), time.Now())
	if err != nil {
		log.Printf("Error inserting record at addRecord \n %s", err)
		return nil, err
	}
	if _, err := res.RowsAffected(); err != nil {
		log.Printf("Error fetching rows \n %s", err)
		return nil, err
	}
	lastInsertID, err := res.LastInsertId()
	if err != nil {
		log.Printf("Error retreiving last insert id \n %s", err)
		return nil, err
	}
	insertRow := uModel.DB.QueryRow("SELECT long_url, short_url, expiry_time, created_at, updated_at from urls where id = ?", lastInsertID)
	newURLRow := URL{}
	err = insertRow.Scan(&newURLRow.LongURL, &newURLRow.ShortURL, &newURLRow.ExpiryTime, &newURLRow.CreatedAt, &newURLRow.UpdatedAt)
	if err != nil {
		log.Printf("Error at row scan at addRecord \n %s", err)
		return nil, err
	}
	return &newURLRow, nil
}

func (uModel URLModel) findOrCreate(longURL, hash string) (*URL, error) {
	urlRow, err := uModel.findURL(longURL)
	if err != nil {
		if err == sql.ErrNoRows {
			urlRow, err := uModel.addRecord(longURL, hash)
			if err != nil {
				return nil, err
			}
			return urlRow, nil
		}
	} else if urlRow.LongURL != "" && !isExpired(urlRow) {
		return &urlRow, nil
	}

	return nil, err

}
