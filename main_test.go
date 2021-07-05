package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"gotest.tools/assert"
)

type ErrorResponse struct {
	Code    string
	Message string
}

func TestValidation(t *testing.T) {

	type fixture struct {
		config Config
		proxy  Proxy
		server *httptest.Server
	}

	setup := func() *fixture {
		s := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				if request.URL.Path == "/petstore.yaml" {
					file, _ := ioutil.ReadFile("resources/petstore.yaml")
					w.Write(file)
				} else {
					w.Write([]byte("hello world"))
				}
			}),
		)

		url, _ := url.Parse(s.URL)
		port, _ := strconv.Atoi(url.Port())

		config := Config{
			ProxyPort:   8080,
			ServicePort: port,
			OpenapiPath: "/petstore.yaml",
		}

		proxy := Proxy{}
		proxy.init(config)

		return &fixture{
			config: config,
			proxy:  proxy,
			server: s,
		}
	}

	t.Run("test operation found", func(t *testing.T) {
		f := setup()

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", f.server.URL+"/pet/findByStatus", nil)

		f.proxy.ServeHTTP(recorder, request)

		assert.Equal(t, 200, recorder.Code)
		assert.Equal(t, "hello world", string(recorder.Body.Bytes()))
	})

	t.Run("test operation not found", func(t *testing.T) {
		f := setup()

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", f.server.URL+"/unknownoperation", nil)

		f.proxy.ServeHTTP(recorder, request)

		assert.Equal(t, 400, recorder.Code)
		var errorMessage ErrorMessage
		err := json.Unmarshal(recorder.Body.Bytes(), &errorMessage)

		assert.NilError(t, err)
		assert.Equal(t, errorMessage.Code, "400")
		assert.Equal(t, errorMessage.Message, "no matching operation was found")
	})
}
