# Go HTTP Test Helper

A helper library to simplify testing of HTTP endpoints in Go.

<!-- mdox-execgen start -->
<!-- GoDoc: ./testhelper.go -->
<!-- mdox-execgen end -->

## Usage Example

```go
import (
	"net/http"
	"testing"
	"github.com/go-chi/chi/v5"
	"github.com/Okja-Engineering/go-service-kit/pkg/testhelper"
)

func TestHomeEndpoint(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	testCases := []testhelper.TestCase{
		{
			Name:           "Test Home Endpoint",
			URL:            "/",
			Method:         "GET",
			CheckBody:      "Hello, World!",
			CheckBodyCount: 1,
			CheckStatus:    200,
		},
	}

	testhelper.Run(t, router, testCases)
}
```

// See tests for more usage patterns.
