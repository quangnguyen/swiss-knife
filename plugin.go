//nolint:all
package swissknife

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

//nolint:all
type Config struct {
	AuthenticationHeader     bool     `json:"authenticationHeader,omitempty"`
	AuthenticationHeaderName string   `json:"headerName,omitempty"`
	BearerHeader             bool     `json:"bearerHeader,omitempty"`
	BearerHeaderName         string   `json:"bearerHeaderName,omitempty"`
	Keys                     []string `json:"keys,omitempty"`
	RemoveHeadersOnSuccess   bool     `json:"removeHeadersOnSuccess,omitempty"`
	EnableLog                bool     `json:"enableLog,omitempty"`
}

//nolint:all
type Response struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

//nolint:all
func CreateConfig() *Config {
	return &Config{
		AuthenticationHeader:     true,
		AuthenticationHeaderName: "X-API-KEY",
		BearerHeader:             true,
		BearerHeaderName:         "Authorization",
		Keys:                     []string{},
		RemoveHeadersOnSuccess:   true,
		EnableLog:                false,
	}
}

//nolint:all
type SwissKnife struct {
	next                     http.Handler
	authenticationHeader     bool
	authenticationHeaderName string
	bearerHeader             bool
	bearerHeaderName         string
	keys                     map[string]struct{}
	removeHeadersOnSuccess   bool
	enableLog                bool
}

//nolint:all
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.EnableLog {
		_, _ = os.Stdout.WriteString(fmt.Sprintf("Creating plugin: %s instance: %+v, ctx: %+v\n", name, *config, ctx))
	}

	// Check for empty keys
	if len(config.Keys) == 0 {
		return nil, errors.New("must specify at least one valid key")
	}

	// Check at least one header is set
	if !config.AuthenticationHeader && !config.BearerHeader {
		return nil, errors.New("at least one header type must be true")
	}

	keysMap := make(map[string]struct{})
	for _, key := range config.Keys {
		keysMap[key] = struct{}{}
	}

	return &SwissKnife{
		next:                     next,
		authenticationHeader:     config.AuthenticationHeader,
		authenticationHeaderName: config.AuthenticationHeaderName,
		bearerHeader:             config.BearerHeader,
		bearerHeaderName:         config.BearerHeaderName,
		keys:                     keysMap,
		removeHeadersOnSuccess:   config.RemoveHeadersOnSuccess,
		enableLog:                config.EnableLog,
	}, nil
}

func contains(key string, validKeys map[string]struct{}) bool {
	_, exists := validKeys[key]
	return exists
}

func bearer(key string, validKeys map[string]struct{}) bool {
	if !strings.HasPrefix(key, "Bearer ") {
		return false
	}
	extractedKey := strings.TrimPrefix(key, "Bearer ")
	return contains(extractedKey, validKeys)
}

func (ka *SwissKnife) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if ka.enableLog {
		_, _ = os.Stdout.WriteString(fmt.Sprintf("Request: %s %s\n", req.Method, req.URL.String()))
	}

	isAuthorized := false

	if ka.authenticationHeader && contains(req.Header.Get(ka.authenticationHeaderName), ka.keys) {
		isAuthorized = true
		if ka.removeHeadersOnSuccess {
			req.Header.Del(ka.authenticationHeaderName)
		}
	} else if ka.bearerHeader && bearer(req.Header.Get(ka.bearerHeaderName), ka.keys) {
		isAuthorized = true
		if ka.removeHeadersOnSuccess {
			req.Header.Del(ka.bearerHeaderName)
		}
	}

	if isAuthorized {
		if ka.enableLog {
			_, _ = os.Stdout.WriteString(fmt.Sprintf("Authorized request: %s %s\n", req.Method, req.URL.String()))
		}
		ka.next.ServeHTTP(rw, req)
		return
	}

	ka.responseError(rw)
}

func (ka *SwissKnife) responseError(rw http.ResponseWriter) {
	response := Response{
		Message:    "Invalid API Key",
		StatusCode: http.StatusForbidden,
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(response.StatusCode)
	if err := json.NewEncoder(rw).Encode(response); err != nil {
		if ka.enableLog {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("Error sending response: %s\n", err.Error()))
		}
	} else {
		if ka.enableLog {
			_, _ = os.Stdout.WriteString(fmt.Sprintf("Response: %d %s\n", response.StatusCode, response.Message))
		}
	}
}
