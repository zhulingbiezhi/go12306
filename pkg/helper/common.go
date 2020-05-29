package helper

import (
	"net/http"
)

type ValidRequest interface {
	ValidParam() error
}

type ParseRequest interface {
	ParseRequest(r *http.Request) error
}
