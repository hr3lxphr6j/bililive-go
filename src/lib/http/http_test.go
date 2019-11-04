package http

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.EqualValues(t, "bar", r.Header.Get("foo"))
		assert.Len(t, r.URL.Query(), 1)
		assert.EqualValues(t, "test", r.URL.Query().Get("q"))
		w.Write([]byte("OK"))
	}))
	b, err := Get(fmt.Sprintf("%s/test", s.URL), map[string]string{"foo": "bar"}, map[string]string{"q": "test"})
	assert.NoError(t, err)
	assert.EqualValues(t, "OK", string(b))
}

func TestPost(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.EqualValues(t, "bar", r.Header.Get("foo"))
		assert.Len(t, r.URL.Query(), 1)
		assert.EqualValues(t, "test", r.URL.Query().Get("q"))
		b, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.EqualValues(t, "hello", string(b))
		w.Write([]byte("OK"))
	}))
	b, err := Post(fmt.Sprintf("%s/test", s.URL), map[string]string{"foo": "bar"}, map[string]string{"q": "test"}, []byte("hello"))
	assert.NoError(t, err)
	assert.EqualValues(t, "OK", string(b))
}

func TestGZipContent(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.EqualValues(t, "bar", r.Header.Get("foo"))
		assert.Len(t, r.URL.Query(), 1)
		assert.EqualValues(t, "test", r.URL.Query().Get("q"))

		w.Header().Add("Content-Encoding", "gzip")
		gw := gzip.NewWriter(w)
		_, err := gw.Write([]byte("hello, gzip!"))
		assert.NoError(t, err)
		assert.NoError(t, gw.Flush())
		assert.NoError(t, gw.Close())
	}))
	b, err := Get(fmt.Sprintf("%s/test", s.URL), map[string]string{"foo": "bar"}, map[string]string{"q": "test"})
	assert.NoError(t, err)
	assert.EqualValues(t, "hello, gzip!", string(b))
}
