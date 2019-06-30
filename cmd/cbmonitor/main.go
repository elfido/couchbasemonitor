package main

import (
	"cbmonitor/internal/config"
	"cbmonitor/internal/monitor"
	"cbmonitor/internal/monitor/stats"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi"
)

// ToDo:
//  - Add prometheus metrics
//  - Create docker file

// ClustersContainer keeps track of statistics of multiple clusters
type ClustersContainer struct {
	clusters map[string]stats.ClusterStats
	mu       sync.RWMutex
}

func NewClustersContainer() ClustersContainer {
	return ClustersContainer{
		clusters: make(map[string]stats.ClusterStats),
	}
}

// Add refreshes the information for a given cluster
func (cc *ClustersContainer) Add(stats stats.ClusterStats) {
	cc.mu.Lock()
	cc.clusters[stats.Name] = stats
	cc.mu.Unlock()
}

// GetAll returns the information of all clusters
func (cc *ClustersContainer) GetAll() []stats.ClusterStats {
	cc.mu.RLock()
	all := make([]stats.ClusterStats, len(cc.clusters))
	ndx := 0
	for _, stats := range cc.clusters {
		all[ndx] = stats
	}
	cc.mu.RUnlock()
	return all
}

func exitOnError(message string, err error) {
	if err != nil {
		log.Printf("%s: %s\n", message, err)
		os.Exit(1)
	}
}

func main() {
	configFile := flag.String("config", "./config.json", "Configuration file path")
	scrapInterval := flag.Duration("interval", 15*time.Second, "Monitoring interval")
	callsTimeout := flag.Duration("timeout", 3*time.Second, "Monitoring call timeout")
	defaultPassword := flag.String("password", "", "Default password (if you don't want to set one in config file)")
	flag.Parse()
	configuration, err := config.NewFileConfiguration(*configFile)
	exitOnError("Cannot read configuration", err)
	log.Printf("Using configuration from file: %s, found %d clusters", *configFile, len(configuration))
	monitors := make([]*monitor.Monitor, len(configuration))
	for i, cluster := range configuration {
		pass := cluster.Credentials.Password
		if pass == "" {
			pass = *defaultPassword
		}
		log.Printf("\t- %s @ %s://%s:%s\n", cluster.Name, cluster.Protocol, cluster.Hostname, cluster.Port)
		monitor, err := monitor.NewMonitor(cluster.Hostname, cluster.Name, cluster.Credentials.Username,
			pass, cluster.Protocol, cluster.Port)
		exitOnError("Cannot create monitor", err)
		monitor.SetTimeout(*callsTimeout)
		monitors[i] = monitor.Build()
	}

	fullClusterStats := NewClustersContainer()
	go func() {
		for {
			log.Println("Processing")
			responses := make(chan monitor.ClusterInfo, len(monitors))
			for _, monitor := range monitors {
				monitor.Check(responses)
			}
			for i := 0; i < len(monitors); i++ {
				resp := <-responses
				if resp.Err == nil {
					fmt.Println(resp.Stats)
					fullClusterStats.Add(resp.Stats)
				} else {
					fmt.Println(resp.Err)
				}
			}
			close(responses)
			time.Sleep(*scrapInterval)
		}
	}()
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		clusters := fullClusterStats.GetAll()
		clustersBytes, _ := json.Marshal(clusters)
		w.Write(clustersBytes)
	})
	http.ListenAndServe(":3000", r)
}
