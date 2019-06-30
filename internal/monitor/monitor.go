package monitor

import (
	"cbmonitor/internal/monitor/stats"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	// ErrInvalidMontiorOptions
	ErrInvalidMontiorOptions = errors.New("invalid monitor options")
)

// Monitor couchbase stats scrapper
type Monitor struct {
	hosts       []string
	clustername string
	protocol    string
	port        string
	username    string
	password    string
	timeout     time.Duration
	client      http.Client
}

// MonitorBuilder monitor creation helper
type MonitorBuilder struct {
	monitor Monitor
}

type ClusterInfo struct {
	Stats stats.ClusterStats
	Err   error
}

// NewMonitor creates a new stats scrapper
func NewMonitor(hostname, name, username, password, protocol, port string) (MonitorBuilder, error) {
	if hostname == "" {
		return MonitorBuilder{}, ErrInvalidMontiorOptions
	}
	return MonitorBuilder{
		monitor: Monitor{
			clustername: name,
			hosts:       []string{hostname},
			username:    username,
			password:    password,
			protocol:    protocol,
			port:        port,
			timeout:     time.Second * 3,
		},
	}, nil
}

// SetTimeout defines the timeout to call couchbase statistics APIs
func (b *MonitorBuilder) SetTimeout(timeout time.Duration) {
	b.monitor.timeout = timeout
}

func (b *MonitorBuilder) initializeClient() {
	transport := &http.Transport{
		DialTLS: nil,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		TLSHandshakeTimeout: 5 * time.Second,
		MaxIdleConns:        2,
		MaxIdleConnsPerHost: 1,
		MaxConnsPerHost:     5,
		// should be tolerant to more than 1 loop
		IdleConnTimeout:       2 * b.monitor.timeout,
		ResponseHeaderTimeout: 2 * time.Second,
	}
	client := http.Client{
		Transport: transport,
		Timeout:   b.monitor.timeout,
	}
	b.monitor.client = client
}

func (b *MonitorBuilder) Build() *Monitor {
	b.initializeClient()
	return &b.monitor
}

func (m *Monitor) Check(responseChannel chan ClusterInfo) {
	baseUrl := fmt.Sprintf("%s://%s", m.protocol, m.hosts[0])
	auth := stats.Auth{m.username, m.password}
	cluster, err := stats.GetPoolInfo(baseUrl, m.port, auth)
	if err == nil {
		fmt.Println(cluster)
		responseChannel <- ClusterInfo{
			Stats: cluster,
			Err:   nil,
		}
	} else {
		responseChannel <- ClusterInfo{
			Stats: stats.ClusterStats{},
			Err:   err,
		}
	}

}
