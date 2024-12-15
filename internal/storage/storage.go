package storage

import "errors"

var (
	ErrUserExists     = errors.New("user exists")
	ErrUserNotFound   = errors.New("user not found")
	ErrSecretNotFound = errors.New("secret not found")
)
