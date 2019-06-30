package stats

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const (
	nodeExpectedHealth = "healthy"

	errCodeHTTPStatus = "[HTTP_STATUS]"
)

var roles = []string{"kv", "index", "n1ql", "fts", "analytics"}

type poolRawNode struct {
	SystemStats struct {
		CPUUtilizationRate float64 `json:"cpu_utilization_rate"`
		SwapTotal          int64   `json:"swap_total"`
		SwapUsed           int64   `json:"swap_used"`
		MemTotalBytes      int64   `json:"mem_total"`
		MemFree            int64   `json:"mem_free"`
	} `json:"systemStats"`
	InterestingStats struct {
		CMDGet                  int64 `json:"cmd_get"`
		CMDGetHits              int64 `json:"get_hits"`
		MemUsedBytes            int64 `json:"mem_used"`
		Ops                     int64 `json:"ops"`
		DocsActualDiskSizeBytes int64 `json:"couch_docs_actual_disk_size"`
		DocsDataSize            int64 `json:"couch_docs_data_size"`
		CurrentItems            int64 `json:"curr_items"`
		CurrentItemsTotal       int64 `json:"curr_items_tot"`
	} `json:"interestingStats"`
	Hostname             string   `json:"hostname"`
	Uptime               string   `json:"uptime"`
	MemoryTotalBytes     int64    `json:"memoryTotal"`
	MemoryFree           int64    `json:"memoryFree"`
	ClusterMembership    string   `json:"clusterMembership"`
	RecoveryType         string   `json:"recoveryType"`
	Status               string   `json:"status"`
	ClusterCompatibility int64    `json:"clusterCompatibility"`
	Version              string   `json:"version"`
	OS                   string   `json:"os"`
	Services             []string `json:"services"`
}

type poolsRawResponse struct {
	StorageTotals struct {
		RAM struct {
			Total             int64 `json:"total"`
			QuotaTotal        int64 `json:"quotaTotal"`
			QuotaUsed         int64 `json:"quotaUsed"`
			Used              int64 `json:"used"`
			UsedByData        int64 `json:"usedByData"`
			QuotaUsedPerNode  int64 `json:"quotaUsedPerNode"`
			QuotaTotalPerNode int64 `json:"quotaTotalPerNode"`
		} `json:"ram"`
		HDD struct {
			Total      int64 `json:"total"`
			QuotaTotal int64 `json:"quotaTotal"`
			Used       int64 `json:"used"`
			UsedByData int64 `json:"usedByData"`
			Free       int64 `json:"free"`
		} `json:"hdd"`
	} `json:"storageTotals"`
	FTSMemoryQuotaMb   int64  `json:"ftsMemoryQuota,omitempty"`
	IndexMemoryQuotaMb int64  `json:"indexMemoryQuota,omitempty"`
	MemoryQuotaMb      int64  `json:"memoryQuota"`
	Name               string `json:"name"`
	Alerts             []struct {
		Message    string `json:"msg"`
		ServerTime string `json:"serverTime"`
	} `json:"alerts"`
	Nodes           []poolRawNode `json:"nodes"`
	RebalanceStatus string        `json:"rebalanceStatus"`
	MaxBucketCount  int64         `json:"maxBucketCount"`
	IndexStatusURL  string        `json:"indexStatusURI"`
	ClusterName     string        `json:"clusterName"`
	Balanced        bool          `json:"balanced"`
}

type ClusterStats struct {
	Name               string  `json:"name"`
	Balanced           bool    `json:"balanced"`
	RebalanceStatus    string  `json:"balanceStatus"`
	FTSMemoryQuotaMb   int64   `json:"ftsMemoryQuota"`
	IndexMemoryQuotaMb int64   `json:"indexMemoryQuota"`
	MemoryQuotaMb      int64   `json:"memoryQuota"`
	RAMTotal           int64   `json:"ramTotal"`
	RAMUsed            int64   `json:"ramUsed"`
	RAMPctUsed         float64 `json:"ramPctUsed"`
	HDTotal            int64   `json:"hdTotal"`
	HDUsed             int64   `json:"hdUsed"`
	HdPctUsed          float64 `json:"hdPctUsed"`
	GetHitRatio        float64 `json:"getHitRatio"`
	AvailableServices  struct {
		KV        int `json:"kv"`
		Index     int `json:"index"`
		Query     int `json:"query"`
		FTS       int `json:"fts"`
		Analytics int `json:"analytics"`
	} `json:"servicesCount"`
	Alerts struct {
		Cluster []struct {
			Message    string `json:"msg"`
			ServerTime string `json:"serverTime"`
		} `json:"cluster"`
		Calculated []string `json:"calculated"`
	} `json:"alerts"`
	Buckets []Bucket `json:"buckets"`
	Nodes   []Node   `json:"node"`
}

type Node struct {
	Hostname          string   `json:"hostname"`
	MemTotalMb        int64    `json:"memTotalMb"`
	MemFreeMb         int64    `json:"memFreeMb"`
	MemUsedPct        float64  `json:"memPctUsed"`
	ClusterMembership string   `json:"clusterMembership"`
	Status            string   `json:"status"`
	Version           string   `json:"version"`
	OS                string   `json:"OS"`
	Services          []string `json:"services"`
	IsKV              bool     `json:"isKV"`
	KVStats           struct {
		GetOps    int64 `json:"getOps"`
		GetHits   int64 `json:"getHits"`
		Ops       int64 `json:"ops"`
		DocsSize  int64 `json:"docsSize"`
		TotalDocs int64 `json:"totalDocs"`
	} `json:"kvStats,omitempty"`
	CPURate float64 `json:"cpuRate"`
}

type nodesSummary struct {
	nodes       []Node
	alerts      []string
	getHitRatio float64
	services    map[string]int
}

func includes(options []string, lookFor string) bool {
	for i, _ := range options {
		if strings.EqualFold(options[i], lookFor) {
			return true
		}
	}
	return false
}

func summarizeNodes(sourceNodes []poolRawNode) nodesSummary {
	nodes := make([]Node, len(sourceNodes))
	var gets int64
	var hits int64
	calculatedAlerts := []string{}
	versions := make(map[string]int)
	compatibility := make(map[int64]int)
	services := make(map[string]int)
	for i, node := range sourceNodes {
		isKV := includes(node.Services, "kv")
		versions[node.Version] = versions[node.Version] + 1
		compatibility[node.ClusterCompatibility] = compatibility[node.ClusterCompatibility] + 1
		for _, service := range node.Services {
			services[service] += 1
		}
		nodes[i] = Node{
			Hostname:          strings.Split(node.Hostname, ":")[0],
			MemTotalMb:        node.MemoryTotalBytes / mbFromBytes,
			MemFreeMb:         node.MemoryFree / mbFromBytes,
			MemUsedPct:        1 - (float64(node.MemoryFree) / float64(node.MemoryTotalBytes)),
			ClusterMembership: node.ClusterMembership,
			Status:            node.Status,
			Version:           node.Version,
			OS:                node.OS,
			Services:          node.Services,
			IsKV:              isKV,
			KVStats: struct {
				GetOps    int64 `json:"getOps"`
				GetHits   int64 `json:"getHits"`
				Ops       int64 `json:"ops"`
				DocsSize  int64 `json:"docsSize"`
				TotalDocs int64 `json:"totalDocs"`
			}{
				node.InterestingStats.CMDGet, node.InterestingStats.CMDGetHits,
				node.InterestingStats.Ops, node.InterestingStats.DocsDataSize,
				node.InterestingStats.CurrentItemsTotal,
			},
			CPURate: node.SystemStats.CPUUtilizationRate,
		}
		gets += nodes[i].KVStats.GetOps
		hits += nodes[i].KVStats.GetHits
	}
	var getHitRatio float64
	if gets > 0 {
		getHitRatio = float64(hits) / float64(gets)
	}
	if len(versions) > 1 {
		versionList := make([]string, 0)
		for version, _ := range versions {
			versionList = append(versionList, version)
		}
		alert := fmt.Sprintf("Multiple Couchbase versions (%d) in cluster: %s", len(versions), strings.Join(versionList, ","))
		calculatedAlerts = append(calculatedAlerts, alert)
	}
	if len(compatibility) > 1 {
		compatibilityList := make([]string, 0)
		for compatibility, _ := range compatibility {
			compatibilityList = append(compatibilityList, strconv.Itoa(int(compatibility)))
		}
		alert := fmt.Sprintf("Multiple Couchbase compatibility modes (%d) in cluster: %s", len(compatibility), strings.Join(compatibilityList, ","))
		calculatedAlerts = append(calculatedAlerts, alert)
	}
	return nodesSummary{
		nodes:       nodes,
		getHitRatio: getHitRatio,
		alerts:      calculatedAlerts,
		services:    services,
	}
}

func (p poolsRawResponse) toClusterStats() ClusterStats {
	calculatedAlerts := []string{}
	summarizedNodes := summarizeNodes(p.Nodes)
	calculatedAlerts = append(calculatedAlerts, summarizedNodes.alerts...)
	ramPctUsed := 0.0
	if p.StorageTotals.RAM.Total > 0 {
		ramPctUsed = float64(p.StorageTotals.RAM.Used / p.StorageTotals.RAM.Total)
	}
	return ClusterStats{
		Name:               p.ClusterName,
		Balanced:           p.Balanced,
		RebalanceStatus:    p.RebalanceStatus,
		FTSMemoryQuotaMb:   p.FTSMemoryQuotaMb,
		IndexMemoryQuotaMb: p.IndexMemoryQuotaMb,
		MemoryQuotaMb:      p.MemoryQuotaMb,
		RAMTotal:           p.StorageTotals.RAM.Total,
		RAMUsed:            p.StorageTotals.RAM.Used,
		RAMPctUsed:         ramPctUsed,
		HDTotal:            p.StorageTotals.HDD.Total,
		HDUsed:             p.StorageTotals.HDD.Used,
		HdPctUsed:          float64(p.StorageTotals.HDD.Used / p.StorageTotals.HDD.Total),
		GetHitRatio:        summarizedNodes.getHitRatio,
		Alerts: struct {
			Cluster []struct {
				Message    string `json:"msg"`
				ServerTime string `json:"serverTime"`
			} `json:"cluster"`
			Calculated []string `json:"calculated"`
		}{
			p.Alerts, calculatedAlerts,
		},
		Buckets: nil,
		Nodes:   summarizedNodes.nodes,
		AvailableServices: struct {
			KV        int `json:"kv"`
			Index     int `json:"index"`
			Query     int `json:"query"`
			FTS       int `json:"fts"`
			Analytics int `json:"analytics"`
		}{
			KV: summarizedNodes.services["kv"], Index: summarizedNodes.services["index"],
			Query: summarizedNodes.services["n1ql"], FTS: summarizedNodes.services["fts"],
			Analytics: summarizedNodes.services["analytics"],
		},
	}
}

func (c ClusterStats) String() string {
	version := c.Nodes[0].Version
	maxCPU := 0.0
	maxMem := 0.0
	totalAlerts := c.Alerts.Cluster
	totalAlerts = append(totalAlerts, c.Alerts.Calculated...)
	alertsCount := len(c.Alerts.Cluster) + len(c.Alerts.Calculated)
	alerts := strings.Join(totalAlerts, "- %s\n")
	for i, _ := range c.Nodes {
		if c.Nodes[i].CPURate > maxCPU {
			maxCPU = c.Nodes[i].CPURate
		}
		if c.Nodes[i].MemUsedPct > maxMem {
			maxMem = c.Nodes[i].MemUsedPct
		}
	}
	maxMem *= 100
	buckets := make([]string, len(c.Buckets))
	for i, bucket := range c.Buckets {
		buckets[i] = bucket.Name
	}
	return fmt.Sprintf("%s - Version: %s\nNodes: %d\tMax CPU: %.1f\tMax Mem used: %.1f\tGet hit/miss ratio: %.1f"+
		"\nBuckets: %s\n"+
		"\nServices:\n- KV: %d"+
		"\nAlerts (%d):\n%s", c.Name, version, len(c.Nodes), maxCPU, maxMem, c.GetHitRatio, strings.Join(buckets, ", "),
		c.AvailableServices.KV, alertsCount, alerts)
}

func GetPoolInfo(baseUrl string, port string, auth Auth) (ClusterStats, error) {
	url := fmt.Sprintf("%s:%s/pools/default", baseUrl, port)
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(auth.Username, auth.Password)
	resp, err := client.Do(req)
	if err != nil {
		return ClusterStats{}, err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return ClusterStats{}, fmt.Errorf("%s invalid status pools API response code: %d", errCodeHTTPStatus,
			resp.StatusCode)
	}
	var poolsResponse poolsRawResponse
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(&poolsResponse)
	resp.Body.Close()
	clusterStats := poolsResponse.toClusterStats()
	bucketsChannel := make(chan bucketsChanResponse)
	go getBuckets(baseUrl, port, auth, bucketsChannel)
	// todo: fetch indices from remote url
	bucketsResponse := <-bucketsChannel
	clusterStats.Buckets = bucketsResponse.buckets
	return clusterStats, nil
}
