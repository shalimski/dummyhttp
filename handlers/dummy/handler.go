package dummy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shalimski/dummyhttp/config"
)

type Handler struct {
	message string
}

type response struct {
	Proto         string            `json:"proto"`
	Host          string            `json:"host"`
	Request       string            `json:"request"`
	Headers       map[string]string `json:"headers"`
	Message       string            `json:"message"`
	Body          string            `json:"body"`
	RemoteAddr    string            `json:"remote_addr"`
	ContentLength int64             `json:"content_length"`
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	resp := response{
		Proto:         r.Proto,
		Host:          r.Host,
		Request:       r.Method + " " + r.RequestURI,
		Headers:       make(map[string]string, len(r.Header)),
		Message:       h.message,
		Body:          "",
		RemoteAddr:    r.RemoteAddr,
		ContentLength: r.ContentLength,
	}

	buf := bytes.Buffer{}
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp.Body = buf.String()
	defer r.Body.Close()

	for k, vv := range r.Header {
		resp.Headers[k] = strings.Join(vv, "; ")
	}

	jsonResp, err := json.MarshalIndent(resp, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(jsonResp)
}

func New(cfg config.HandlerConfig) *Handler {
	return &Handler{
		message: cfg.Message,
	}
}
