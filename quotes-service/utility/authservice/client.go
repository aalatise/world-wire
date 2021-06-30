package authservice

import (
	"net/http"
)

type Client struct {
	HTTP *http.Client
}
