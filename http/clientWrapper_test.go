package http

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

const testUrl = "/testurl"

func TestNewHTTPClientWrapper(t *testing.T) {
	t.Parallel()

	t.Run("invalid timeout should error", func(t *testing.T) {
		client, err := NewHTTPClientWrapper(time.Second - time.Nanosecond)

		assert.True(t, check.IfNil(client))
		assert.True(t, errors.Is(err, errInvalidValue))
	})
	t.Run("should work", func(t *testing.T) {
		client, err := NewHTTPClientWrapper(time.Second)

		assert.False(t, check.IfNil(client))
		assert.Nil(t, err)
	})
}

func TestHttpClientWrapper_CallGetRestEndPoint(t *testing.T) {
	t.Parallel()

	t.Run("nil context should error", func(t *testing.T) {
		client, _ := NewHTTPClientWrapper(time.Second)

		buff, err := client.CallGetRestEndPoint(nil, "")

		assert.Nil(t, buff)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "nil Context"))
	})
	t.Run("invalid url should error", func(t *testing.T) {
		client, _ := NewHTTPClientWrapper(time.Second)

		buff, err := client.CallGetRestEndPoint(context.Background(), "invalid url")

		assert.Nil(t, buff)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "unsupported protocol scheme"))
	})
	t.Run("should work", func(t *testing.T) {
		buffToSend := []byte("buffer to send")

		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, http.MethodGet)
			assert.Equal(t, testUrl, r.URL.Path)
			assert.Equal(t, applicationType, r.Header.Get("Accept"))
			assert.Equal(t, userAgent, r.Header.Get("User-Agent"))

			_, err := w.Write(buffToSend)
			assert.Nil(t, err)
		}))

		client, _ := NewHTTPClientWrapper(time.Second)

		buff, err := client.CallGetRestEndPoint(context.Background(), svr.URL+testUrl)

		assert.Nil(t, err)
		assert.Equal(t, buffToSend, buff)
	})
}

func TestHttpClientWrapper_CallPostRestEndPoint(t *testing.T) {
	t.Parallel()

	t.Run("json.Marshal fails should error", func(t *testing.T) {
		client, _ := NewHTTPClientWrapper(time.Second)

		// we can not marshal a function pointer
		err := client.CallPostRestEndPoint(context.Background(), "", func() {})

		assert.NotNil(t, err)
		assert.Equal(t, "*json.UnsupportedTypeError", fmt.Sprintf("%T", err))
	})
	t.Run("nil context should error", func(t *testing.T) {
		client, _ := NewHTTPClientWrapper(time.Second)

		err := client.CallPostRestEndPoint(nil, "", "data")

		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "nil Context"))
	})
	t.Run("invalid url should error", func(t *testing.T) {
		client, _ := NewHTTPClientWrapper(time.Second)

		err := client.CallPostRestEndPoint(context.Background(), "invalid url", "test")

		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "unsupported protocol scheme"))
	})
	t.Run("should work", func(t *testing.T) {
		expectedBuff := []byte(`{"fielda":"a","fieldb":1}`)

		var result []byte

		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, http.MethodPost)
			assert.Equal(t, testUrl, r.URL.Path)
			assert.Equal(t, applicationType, r.Header.Get("Accept"))
			assert.Equal(t, userAgent, r.Header.Get("User-Agent"))
			assert.Equal(t, applicationType, r.Header.Get("Content-Type"))

			defer func() {
				_ = r.Body.Close()
			}()

			buff, err := ioutil.ReadAll(r.Body)
			assert.Nil(t, err)
			result = buff
		}))

		client, _ := NewHTTPClientWrapper(time.Second)

		data := struct {
			FieldA string `json:"fielda"`
			FieldB int    `json:"fieldb"`
		}{
			FieldA: "a",
			FieldB: 1,
		}

		err := client.CallPostRestEndPoint(context.Background(), svr.URL+testUrl, &data)

		assert.Nil(t, err)
		assert.Equal(t, string(expectedBuff), string(result))
	})
}
