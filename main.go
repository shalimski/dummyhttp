package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	listen  = flag.String("l", ":8080", "address to listen on")
	message = flag.String("m", "hello, world", "message to return")
	help    = flag.Bool("h", false, "show help")
)

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

func main() {
	slog.Info("dummy-http")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	const commonTimeout = 5 * time.Second
	srv := http.Server{
		Addr:              *listen,
		Handler:           handler(),
		ReadHeaderTimeout: commonTimeout,
		WriteTimeout:      commonTimeout,
		IdleTimeout:       commonTimeout,
	}
	slog.Info("starting server", "listen", *listen)

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("shutdown", "error", err.Error())
		}
	}()

	if err := srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
		} else {
			slog.Info("server stopped")
		}
	}
}

func handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := response{
			Proto:         r.Proto,
			Host:          r.Host,
			Request:       r.Method + " " + r.RequestURI,
			Headers:       make(map[string]string, len(r.Header)),
			Message:       *message,
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
}
