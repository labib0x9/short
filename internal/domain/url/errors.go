package url

import "errors"

var (
	ErrUrlNotFound        = errors.New("URL not found")
	ErrUrlAlreadyExists   = errors.New("URL already exists")
	ErrInvalidUrl         = errors.New("invalid URL format")
	ErrShortCodeExpired   = errors.New("short code has expired")
	ErrShortCodeCollision = errors.New("short code collision detected")
	ErrDatabaseConnection = errors.New("database connection error")
	ErrDatabaseQuery      = errors.New("database query error")
	ErrDatabaseInsert     = errors.New("database insert error")
	ErrDatabaseUpdate     = errors.New("database update error")
	ErrDatabaseDelete     = errors.New("database delete error")
)
