package services

import "errors"

var ErrPortfolioNotFound = errors.New("portfolio not found")
var ErrInvalidUserID = errors.New("invalid user id")
var ErrInvalidAllocation = errors.New("invalid allocation")
