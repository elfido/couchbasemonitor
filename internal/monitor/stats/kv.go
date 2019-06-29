package stats

// KV Key-value API statistics
type KV struct {
}

type kvRaw struct {
	HitRatio  int `json:"hit_ratio"`
	QueueSize int `json:"ep"`
}

func GetKVStats() (KV, error) {
	return KV{}, nil
}
