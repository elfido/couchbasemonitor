package config

import (
	"encoding/json"
	"io/ioutil"
)

// Cluster couchbase cluster information
type Cluster struct {
	Credentials Auth
	Name string
	Hostname string
}

// Auth simple authentication
type Auth struct {
	Username string `json:"user"`
	Password string `json:"password"`
}

type clusterInfo struct {
	Name string `json:"name"`
	Hostname string `json:"hostname"`
}

type configFile struct {
	DefaultAuth Auth
	Clusters []clusterInfo
}

// NewFileConfiguration extracts the configuration of multiple clusters from a given file
func NewFileConfiguration(filename string) ([]Cluster, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return []Cluster{}, err
	}
	var fileContent configFile
	err = json.Unmarshal(bytes, &fileContent)
	if err != nil {
		return []Cluster{}, err
	}
	clusters := make([]Cluster, len(fileContent.Clusters))
	for i, clusterConfig := range fileContent.Clusters {
		cluster := Cluster{
			Credentials: Auth{
				Username: fileContent.DefaultAuth.Username,
				Password: fileContent.DefaultAuth.Password,
			},
			Name:        clusterConfig.Name,
			Hostname:    clusterConfig.Hostname,
		}
		clusters[i] = cluster
	}
	return clusters, nil
}