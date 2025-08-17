# Creating the API

To create an API instance, initialize a Base struct with the necessary information:

## Functions

`StartServer(port int, router chi.Router, timeout time.Duration)`

Starts the API server on the specified port, using the provided chi.Router to handle incoming requests. Configures the server with the specified timeout duration for read and write operations.

`ReturnJSON(resp http.ResponseWriter, data interface{})`

Serializes the provided data as JSON and writes it to the response.

`ReturnText(resp http.ResponseWriter, msg string)`

Writes the provided msg as plain text to the response.

`ReturnErrorJSON(resp http.ResponseWriter, err error)`

Returns an error JSON response with the specified err message.

`ReturnOKJSON(resp http.ResponseWriter)`

Returns an OK JSON response.

## Usage

```go
type MyAPI struct {
  *api.Base
  // Add extra fields as per your implementation
}

router := chi.NewRouter()
api := MyAPI{
  api.NewBase(serviceName, version, buildInfo, healthy),
}

api.AddHealthEndpoint(router, "health")
api.AddStatusEndpoint(router, "status")
```

// See tests for more usage patterns.
