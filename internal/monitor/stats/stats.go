package stats

import "net/http"

var (
	client = http.Client{}
)

const (
	kbFromBytes = 1024
	mbFromBytes = 1024 * 1024
)

type Auth struct {
	Username string
	Password string
}

// SetGlobalClient defines the default HTTP client that will be used to call the monitoring APIs
func SetGlobalClient(apiClient http.Client) {
	client = apiClient
}
