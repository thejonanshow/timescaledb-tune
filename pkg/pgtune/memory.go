package pgtune

import (
	"fmt"
	"math"
	"runtime"

	"github.com/timescale/timescaledb-tune/internal/parse"
)

// Keys in the conf file that are tuned related to memory
const (
	SharedBuffersKey      = "shared_buffers"
	EffectiveCacheKey     = "effective_cache_size"
	MaintenanceWorkMemKey = "maintenance_work_mem"
	WorkMemKey            = "work_mem"

	sharedBuffersWindows = 512 * parse.Megabyte
)

// MemoryKeys is an array of keys that are tunable for memory
var MemoryKeys = []string{
	SharedBuffersKey,
	EffectiveCacheKey,
	MaintenanceWorkMemKey,
	WorkMemKey,
}

// MemoryRecommender gives recommendations for ParallelKeys based on system resources
type MemoryRecommender struct {
	totalMem uint64
	cpus     int
}

// NewMemoryRecommender returns a MemoryRecommender that recommends based on the given
// number of cpus and system memory
func NewMemoryRecommender(totalMemory uint64, cpus int) *MemoryRecommender {
	return &MemoryRecommender{totalMemory, cpus}
}

// IsAvailable returns whether this Recommender is usable given the system resources. Always true.
func (r *MemoryRecommender) IsAvailable() bool {
	return true
}

// Recommend returns the recommended PostgreSQL formatted value for the conf
// file for a given key.
func (r *MemoryRecommender) Recommend(key string) string {
	var val string
	if key == SharedBuffersKey {
		if runtime.GOOS == osWindows {
			val = parse.BytesToPGFormat(sharedBuffersWindows)
		} else {
			val = parse.BytesToPGFormat(r.totalMem / 4)
		}
	} else if key == EffectiveCacheKey {
		val = parse.BytesToPGFormat((r.totalMem * 3) / 4)
	} else if key == MaintenanceWorkMemKey {
		temp := (float64(r.totalMem) / float64(parse.Gigabyte)) * (128.0 * float64(parse.Megabyte))
		if temp > (2 * parse.Gigabyte) {
			temp = 2 * parse.Gigabyte
		}
		val = parse.BytesToPGFormat(uint64(temp))
	} else if key == WorkMemKey {
		if runtime.GOOS == osWindows {
			val = r.recommendWindows()
		} else {
			cpuFactor := math.Round(float64(r.cpus) / 2.0)
			temp := (float64(r.totalMem) / float64(parse.Gigabyte)) * (6.4 * float64(parse.Megabyte)) / cpuFactor
			val = parse.BytesToPGFormat(uint64(temp))
		}
	} else {
		panic(fmt.Sprintf("unknown key: %s", key))
	}
	return val
}

func (r *MemoryRecommender) recommendWindows() string {
	cpuFactor := math.Round(float64(r.cpus) / 2.0)
	if r.totalMem <= 2*parse.Gigabyte {
		temp := (float64(r.totalMem) / float64(parse.Gigabyte)) * (6.4 * float64(parse.Megabyte)) / cpuFactor
		return parse.BytesToPGFormat(uint64(temp))
	}
	base := 2.0 * 6.4 * float64(parse.Megabyte)
	temp := ((float64(r.totalMem)/float64(parse.Gigabyte)-2)*(8.53336*float64(parse.Megabyte)) + base) / cpuFactor
	return parse.BytesToPGFormat(uint64(temp))
}