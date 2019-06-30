package stats

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type bucketRaw struct {
	Name          string `json:"name"`
	BucketType    string `json:"bucketType"`
	ReplicaNumber int    `json:"replicaNumber"`
	BasicStats    struct {
		QuotaPercentUsed float64 `json:"quotaPercentUsed"`
		OpsPerSec        int     `json:"opsPerSec"`
		DiskFetches      int     `json:"diskFetches"`
		ItemCount        int     `json:"itemCount"`
		DiskUsed         int64   `json:"diskUsed"`
		DataUsed         int64   `json:"dataUsed"`
		MemUsed          int64   `json:"memUsed"`
		StorageTotals    struct {
			RAM struct {
				Total            int64 `json:"total"`
				QuotaTotal       int64 `json:"quotaTotal"`
				QuotaUsed        int64 `json:"quotaUsed"`
				Used             int64 `json:"used"`
				QuotaUsedPerNode int64 `json:"quotaTotalPerNode"`
			} `json:"ram"`
			HDD struct {
				Total      int64 `json:"total"`
				QuotaTotal int64 `json:"quotaTotal"`
				Used       int64 `json:"used"`
				Free       int64 `json:"free"`
			}
		} `json:"storageTotals"`
	} `json:"basicStats"`
}

type Bucket struct {
	Name          string  `json:"name"`
	BucketType    string  `json:"bucketType"`
	ReplicaNumber int     `json:"replicaNumber"`
	OpsPerSec     int     `json:"opsPerSec"`
	DiskFetches   int     `json:"diskFetches"`
	ItemCount     int     `json:"itemCount"`
	MemUsedMb     int64   `json:"memUsedMb"`
	QuotaPctUsed  float64 `json:"quotaPctUsed"`
	DiskUsedMb    int64   `json:"diskUsedMb"`
	RAMTotalMB    int     `json:"ramTotalMb"`
	RAMUsedMB     int     `json:"ramUsedMb"`
	RAMFreeMb     int     `json:"ramFreeMb"`
	RAMUsedPct    float64 `json:"ramUsedPct"`
	HDDTotalMb    int     `json:"hddTotalMb"`
	HDDUsedMb     int     `json:"hdUsedMb"`
	HDDFreeMb     int     `json:"hdFreeMb"`
	HDDUsedPct    float64 `json:"hdUsedPct"`
}

type bucketsChanResponse struct {
	buckets []Bucket
	err     error
}

func (br bucketRaw) toBucketSumamry() Bucket {
	freeRam := br.BasicStats.StorageTotals.RAM.Total - br.BasicStats.StorageTotals.RAM.Used
	return Bucket{
		Name:          br.Name,
		BucketType:    br.BucketType,
		ReplicaNumber: br.ReplicaNumber,
		OpsPerSec:     br.BasicStats.OpsPerSec,
		DiskFetches:   br.BasicStats.DiskFetches,
		ItemCount:     br.BasicStats.ItemCount,
		MemUsedMb:     br.BasicStats.MemUsed * mbFromBytes,
		QuotaPctUsed:  br.BasicStats.QuotaPercentUsed,
		DiskUsedMb:    br.BasicStats.DiskUsed,
		RAMTotalMB:    int(br.BasicStats.StorageTotals.RAM.Total * int64(mbFromBytes)),
		RAMUsedMB:     int(br.BasicStats.StorageTotals.RAM.Used * int64(mbFromBytes)),
		RAMFreeMb:     int(freeRam * int64(mbFromBytes)),
		RAMUsedPct:    0,
		HDDTotalMb:    0,
		HDDUsedMb:     0,
		HDDFreeMb:     0,
		HDDUsedPct:    0,
	}
}

func getBuckets(baseUrl, port string, auth Auth, responseChannel chan bucketsChanResponse) {
	url := fmt.Sprintf("%s:%s/pools/default/buckets?basic_stats=true&skipMap=true", baseUrl, port)
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(auth.Username, auth.Password)
	resp, err := client.Do(req)
	if err != nil {
		responseChannel <- bucketsChanResponse{
			buckets: []Bucket{},
			err:     err,
		}
		return
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		responseChannel <- bucketsChanResponse{
			buckets: []Bucket{},
			err: fmt.Errorf("%s invalid status buckets API response code: %d", errCodeHTTPStatus,
				resp.StatusCode),
		}
		return
	}
	var bucketsRaw []bucketRaw
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(&bucketsRaw)
	resp.Body.Close()
	buckets := make([]Bucket, len(bucketsRaw))
	for i, bucket := range bucketsRaw {
		buckets[i] = bucket.toBucketSumamry()
	}
	responseChannel <- bucketsChanResponse{
		buckets: buckets,
		err:     nil,
	}
}
