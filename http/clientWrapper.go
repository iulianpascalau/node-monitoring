package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("http/client")

const (
	minRequestTimeout = time.Second
	userAgent         = "Elrond Node Monitoring / 1.0.0 <Requesting data from api>"
	applicationType   = "application/json"
)

type httpClientWrapper struct {
	httpClient *http.Client
}

// NewHTTPClientWrapper creates an instance of httpClient which is a wrapper for http.Client
func NewHTTPClientWrapper(requestTimeout time.Duration) (*httpClientWrapper, error) {
	if requestTimeout < minRequestTimeout {
		return nil, fmt.Errorf("%w, provided: %v, minimum: %v", errInvalidValue, requestTimeout, minRequestTimeout)
	}

	httpClient := http.DefaultClient
	httpClient.Timeout = requestTimeout

	return &httpClientWrapper{
		httpClient: httpClient,
	}, nil
}

// CallGetRestEndPoint calls an external end point
func (hcw *httpClientWrapper) CallGetRestEndPoint(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	applyGetHeaders(req)
	resp, err := hcw.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		errNotCritical := resp.Body.Close()
		if errNotCritical != nil {
			log.Warn("base process GET: close body", "error", errNotCritical.Error())
		}
	}()

	return ioutil.ReadAll(resp.Body)
}

// CallPostRestEndPoint calls an external end point
func (hcw *httpClientWrapper) CallPostRestEndPoint(ctx context.Context, url string, data interface{}) error {
	buff, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buff))
	if err != nil {
		return err
	}

	applyPostHeaders(req)
	resp, err := hcw.httpClient.Do(req)
	if err != nil {
		return err
	}

	errNotCritical := resp.Body.Close()
	if errNotCritical != nil {
		log.Warn("base process GET: close body", "error", errNotCritical.Error())
	}

	return nil
}

func applyGetHeaders(request *http.Request) {
	request.Header.Set("Accept", applicationType)
	request.Header.Set("User-Agent", userAgent)
}

func applyPostHeaders(request *http.Request) {
	applyGetHeaders(request)
	request.Header.Set("Content-Type", applicationType)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hcw *httpClientWrapper) IsInterfaceNil() bool {
	return hcw == nil
}
