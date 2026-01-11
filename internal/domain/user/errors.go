package user

import "errors"

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInactiveAccount = errors.New("account is inactive or banned")
