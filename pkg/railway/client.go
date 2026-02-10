package railway

import (
	"net/http"
)

type Client struct {
	client        *http.Client
	environmentID string
	token         string
}

func NewClient(environmentID, token string) (*Client, error) {
	c := new(http.Client)
	return &Client{
		client:        c,
		environmentID: environmentID,
		token:         token,
	}, nil
}
