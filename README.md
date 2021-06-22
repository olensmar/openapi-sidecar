# openapi-sidecar

A simple HTTP proxy for validating incoming requests against an OpenAPI definition and returning nice errors if they
aren't compliant with any defined operation or its parameters/schema.

Useful if you

* don't want to write all that logic yourself
* want to guard your service from unwanted illegal requests

Supports both OpenAPI 2.0 and 3.X definitions.

## Example

The included [deployment.yaml](k8s/deployment.yaml) will deploy the swagger3 petstore service with the openapi-sidecar
configured with port-rewriting (see below). Add it to your cluster with

```shell
kubectl apply -f k8s/deployment.yaml
```

Use port-forwarding to make the service available on port 8001 on your local machine:

```shell
kubectl port-forward service/petstore 8001:80
Forwarding from 127.0.0.1:8001 -> 8000
Forwarding from [::1]:8001 -> 8000
```

(you should be able to access Swagger-UI at http://localhost:8001 in your browser now)

Try sending a valid request to the service with curl:

```shell
curl -X GET "http://localhost:8001/api/v3/pet/findByStatus?status=sold" -H  "accept: application/json"
[{"id":5,"category":{"id":1,"name":"Dogs"},"name":"Dog 2","photoUrls":["url1","url2"],"tags":[{"id":1,"name":"tag2"},{"id":2,"name":"tag3"}],"status":"sold"}]
```

Now try sending an invalid request with an incorrect status value:

```shell
 curl -X GET "http://localhost:8001/api/v3/pet/findByStatus?status=something" -H  "accept: application/json"          
{"code":"400","message":"parameter \"status\" in query has an error: value is not one of the allowed values\nSchema:\n  {\n    \"default\": \"available\",\n    \"enum\": [\n      \"available\",\n      \"pending\",\n      \"sold\"\n    ],\n    \"type\": \"string\"\n  }\n\nValue:\n  \"something\"\n | "}
```

Use [jq](https://stedolan.github.io/jq/) to extract the message:

```shell
curl -X GET "http://localhost:8001/api/v3/pet/findByStatus?status=something" -H  "accept: application/json" | jq -r '.message'
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   302  100   302    0     0  15894      0 --:--:-- --:--:-- --:--:-- 15894
parameter "status" in query has an error: value is not one of the allowed values
Schema:
  {
    "default": "available",
    "enum": [
      "available",
      "pending",
      "sold"
    ],
    "type": "string"
  }

Value:
  "something"
 | 
```

### Regular Proxy

To use the openapi-sidecar, add it to your pod/deployment as follows:

```yaml
      containers:
        - .. your service container
        - name: openapi-proxy
          image: kubeshop/openapi-sidecar:latest
          ports:
            - containerPort: <port you want to expose>
          env:
            - name: PROXY_PORT
              value: "<port you want to proxy to listen on"
            - name: SERVICE_PORT
              value: "<port your target service is listening on>"
            - name: OPENAPI_PATH
              value: "<url or path to the openapi definition to use for validation>"
```

If you specify a path (and not a complete URL) for the OpenAPI definition it will be retrieved from the target service
at the specified SERVICE_PORT.

The following example taken from the included [deployment.yaml](k8s/deployment.yaml) is for the petstore3 service (which
incidentally works with all the default values):

```yaml
      containers:
        - name: service
          image: swaggerapi/petstore3:unstable
          ports:
            - containerPort: 8080
        - name: proxy
          image: kubeshop/openapi-sidecar:latest
          ports:
            - containerPort: 8000
```

In this case clients would use port 8000 instead of 8080 to get request-validation functionality.

### With port-rewriting

If you for some reason can't (or don't want to) change the port of the target service (or the service client) you can
use the included openapi-sidecar-init docker image to dynamically rewrite the ports of the pod.

Adding this to the example above:

```yaml
      initContainers:
        - name: init-networking
          image: kubeshop/openapi-sidecar-init:latest
          securityContext:
            capabilities:
              add:
                - NET_ADMIN
            privileged: true
          env:
            - name: SOURCE_PORT
              value: "8080"
            - name: TARGET_PORT
              value: "8000"
      containers:
        - name: service
          image: swaggerapi/petstore3:unstable
          ports:
            - containerPort: 8080
        - name: proxy
          image: olensmar/openapi-sidecar:latest
          ports:
            - containerPort: 8000
```

The init-container will rewrite the iptable of the pod as to redirect traffic going to the service on port 8080 to go to
port 8000 instead (which will then proxy back to the "real" 8080).

## Kudos

To Venil Noronha's https://venilnoronha.io/hand-crafting-a-sidecar-proxy-and-demystifying-istio article, which provided
much of the boilerplate and insight needed for this project
