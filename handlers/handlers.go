package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/shalimski/dummyhttp/config"
	"github.com/shalimski/dummyhttp/handlers/dummy"
)

func New(mode string, config config.HandlerConfig) (http.HandlerFunc, error) {
	switch mode {
	case "dummy":
		h := dummy.New(config)
		return h.Handle, nil
	case "static":
		return nil, errors.New("not implemented")
	case "openapi":
		return nil, errors.New("not implemented")
	default:
		return nil, fmt.Errorf("unknown mode: %s", mode)
	}
}
