package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
)

type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Config struct {
	ProxyPort   int    `env:"PROXY_PORT,required"`
	ServicePort int    `env:"SERVICE_PORT,required"`
	OpenapiPath string `env:"OPENAPI_PATH,required"`
}

type Proxy struct {
	router routers.Router
	config Config
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	err := p.validateRequest(req)
	if err != nil {

		msg := ErrorMessage{
			Code:    "400",
			Message: err.Error(),
		}
		errorMessage, err := json.Marshal(msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-type", "application/json; charset=utf-8")
		w.WriteHeader(400)
		w.Write(errorMessage)
		return
	}

	// Forward the HTTP request to the destination service.
	res, err := p.forwardRequest(req)

	// Notify the client if there was an error while forwarding the request.
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// write the response back to the client.
	p.writeResponse(w, res)
}

func (p *Proxy) forwardRequest(req *http.Request) (*http.Response, error) {
	// Prepare the destination endpoint to forward the request to.
	proxyUrl := fmt.Sprintf("http://127.0.0.1:%d%s", p.config.ServicePort, req.URL.Path)

	// Create an HTTP client and a proxy request based on the original request.
	httpClient := http.Client{}
	proxyReq, err := http.NewRequest(req.Method, proxyUrl, req.Body)
	res, err := httpClient.Do(proxyReq)

	return res, err
}

func (p *Proxy) writeResponse(w http.ResponseWriter, res *http.Response) {
	// Copy all the header values from the response.
	for name, values := range res.Header {
		w.Header()[name] = values
	}

	// Set a special header to notify that the proxy actually serviced the request.
	w.Header().Set("Server", "openapi-sidecar")

	// Set the status code returned by the destination service.
	w.WriteHeader(res.StatusCode)

	// Copy the contents from the response body.
	io.Copy(w, res.Body)

	// Finish the request.
	res.Body.Close()
}

func (p *Proxy) Init(config Config) error {
	ctx := context.Background()
	loader := &openapi3.Loader{Context: ctx}
	urlpath := config.OpenapiPath

	if !strings.HasPrefix(strings.ToLower(urlpath), "http://") && !strings.HasPrefix(strings.ToLower(urlpath), "https://") {
		urlpath = fmt.Sprintf("http://127.0.0.1:%d%s", config.ServicePort, config.OpenapiPath)
	}

	openapiUrl, err := url.Parse(urlpath)
	if err != nil {
		return err
	}

	doc, err := loader.LoadFromURI(openapiUrl)
	if err != nil {
		return err
	}

	err = doc.Validate(ctx)
	if err != nil {
		return err
	}

	p.router, err = legacyrouter.NewRouter(doc)
	p.config = config
	return err
}

func (p *Proxy) validateRequest(httpReq *http.Request) error {
	ctx := context.Background()

	// Find route
	route, pathParams, err := p.router.FindRoute(httpReq)
	if err != nil {
		fmt.Errorf("error finding route: %v", err)
		return err
	}

	// Validate request
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
		Options: &openapi3filter.Options{
			MultiError: true,
			AuthenticationFunc: func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
				return nil
			},
		},
	}

	return openapi3filter.ValidateRequest(ctx, requestValidationInput)
}
