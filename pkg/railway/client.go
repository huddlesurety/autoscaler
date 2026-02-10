package railway

import (
	"net/http"

	"github.com/huddlesurety/autoscaler/internal/config"
)

type Client struct {
	cfg    *config.Config
	client *http.Client
}

func NewClient(cfg *config.Config) (*Client, error) {
	c := new(http.Client)
	return &Client{
		cfg:    cfg,
		client: c,
	}, nil
}
