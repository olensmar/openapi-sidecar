package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/kubeshop/openapi-sidecar/pkg/proxy"
	"github.com/sethvargo/go-envconfig"
)

func main() {
	ctx := context.Background()
	var config proxy.Config
	if err := envconfig.Process(ctx, &config); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Initializing OpenAPI Sidecar from : %s, proxying port %d => %d\n", config.OpenapiPath, config.ProxyPort, config.ServicePort)

	proxy := proxy.Proxy{}
	err := proxy.Init(config)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Print("Initialized OpenAPI Sidecar...")
		http.ListenAndServe(fmt.Sprintf(":%d", config.ProxyPort), &proxy)
	}
}
