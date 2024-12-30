package dummy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shalimski/dummyhttp/config"
)

func TestHandler_New(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		config  config.HandlerConfig
		wantMsg string
	}{
		{
			name: "basic configuration",
			config: config.HandlerConfig{
				Message: "test message",
			},
			wantMsg: "test message",
		},
		{
			name:    "empty message",
			config:  config.HandlerConfig{},
			wantMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(tt.config)
			assert.NotNil(t, h)
			assert.Equal(t, tt.wantMsg, h.message)
		})
	}
}

func TestHandler_Handle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		message       string
		method        string
		path          string
		headers       map[string]string
		body          string
		expectedCode  int
		expectedProto string
		validateFunc  func(*testing.T, response)
	}{
		{
			name:    "basic GET request",
			message: "hello test",
			method:  "GET",
			path:    "/test",
			headers: map[string]string{
				"X-Test-Header": "test-value",
			},
			expectedCode:  http.StatusOK,
			expectedProto: "HTTP/1.1",
			validateFunc: func(t *testing.T, resp response) {
				assert.Equal(t, "hello test", resp.Message)
				assert.Equal(t, "GET /test", resp.Request)
				assert.Equal(t, "test-value", resp.Headers["X-Test-Header"])
				assert.Empty(t, resp.Body)
			},
		},
		{
			name:    "POST request with body",
			message: "hello post",
			method:  "POST",
			path:    "/api/data",
			body:    `{"key": "value"}`,
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			expectedCode:  http.StatusOK,
			expectedProto: "HTTP/1.1",
			validateFunc: func(t *testing.T, resp response) {
				assert.Equal(t, "hello post", resp.Message)
				assert.Equal(t, "POST /api/data", resp.Request)
				assert.Equal(t, "application/json", resp.Headers["Content-Type"])
				assert.Equal(t, `{"key": "value"}`, resp.Body)
				assert.Equal(t, int64(len(`{"key": "value"}`)), resp.ContentLength)
			},
		},
		{
			name:    "request with multiple header values",
			message: "test headers",
			method:  "GET",
			path:    "/headers",
			headers: map[string]string{
				"Accept": "application/json, text/plain",
			},
			expectedCode:  http.StatusOK,
			expectedProto: "HTTP/1.1",
			validateFunc: func(t *testing.T, resp response) {
				assert.Equal(t, "application/json, text/plain", resp.Headers["Accept"])
			},
		},
		{
			name:          "request with empty body",
			message:       "empty body",
			method:        "PUT",
			path:          "/empty",
			expectedCode:  http.StatusOK,
			expectedProto: "HTTP/1.1",
			validateFunc: func(t *testing.T, resp response) {
				assert.Empty(t, resp.Body)
				assert.Equal(t, int64(0), resp.ContentLength)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(config.HandlerConfig{Message: tt.message})

			body := bytes.NewBufferString(tt.body)
			req := httptest.NewRequest(tt.method, tt.path, body)

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			rr := httptest.NewRecorder()

			h.Handle(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			var resp response
			err := json.NewDecoder(rr.Body).Decode(&resp)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedProto, resp.Proto)
			assert.Equal(t, req.Host, resp.Host)
			assert.NotEmpty(t, resp.RemoteAddr)

			// Выполняем специфичные для теста проверки
			if tt.validateFunc != nil {
				tt.validateFunc(t, resp)
			}
		})
	}
}

func TestHandler_HandleErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		bodyReader   io.Reader
		expectedCode int
	}{
		{
			name:         "error reading body",
			bodyReader:   &errorReader{},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(config.HandlerConfig{Message: "test"})

			req := httptest.NewRequest(http.MethodPost, "/test", tt.bodyReader)
			rr := httptest.NewRecorder()

			h.Handle(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
		})
	}
}

type errorReader struct{}

func (e *errorReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestHandler_LargeBody(t *testing.T) {
	largeBody := make([]byte, 1<<20) // 1 MB
	for i := range largeBody {
		largeBody[i] = 'a'
	}

	h := New(config.HandlerConfig{Message: "test"})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(largeBody))
	rr := httptest.NewRecorder()

	h.Handle(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp response
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, string(largeBody), resp.Body)
	assert.Equal(t, int64(len(largeBody)), resp.ContentLength)
}

func TestHandler_ConcurrentRequests(t *testing.T) {
	h := New(config.HandlerConfig{Message: "test"})

	concurrentRequests := 10
	done := make(chan bool)

	for i := range concurrentRequests {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/test"+strconv.Itoa(i), nil)
			rr := httptest.NewRecorder()

			h.Handle(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			var resp response
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, "test", resp.Message)

			done <- true
		}()
	}

	for i := 0; i < concurrentRequests; i++ {
		<-done
	}
}
