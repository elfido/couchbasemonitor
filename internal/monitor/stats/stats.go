package stats

import "net/http"

var (
	client = http.Client{}
)

type Auth struct {
	Username string
	Password string
}

// SetGlobalClient defines the default HTTP client that will be used to call the monitoring APIs
func SetGlobalClient(apiClient http.Client) {
	client = apiClient
}
