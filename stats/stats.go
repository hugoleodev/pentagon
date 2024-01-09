package stats

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

type MemInfo struct {
	MemTotal     uint64 `json:"mem_total"`
	MemFree      uint64 `json:"mem_free"`
	MemAvailable uint64 `json:"mem_available"`
	Buffers      uint64 `json:"buffers"`
	Cached       uint64 `json:"cached"`
	SwapCached   uint64 `json:"swap_cached"`
	Active       uint64 `json:"active"`
	Inactive     uint64 `json:"inactive"`
	SwapTotal    uint64 `json:"swap_total"`
	SwapFree     uint64 `json:"swap_free"`
	Shmem        uint64 `json:"shmem"`
	VmallocTotal uint64 `json:"vmalloc_total"`
	VmallocUsed  uint64 `json:"vmalloc_used"`
	VmallocChunk uint64 `json:"vmalloc_chunk"`
}

type Disk struct {
	All        uint64 `json:"all"`
	Used       uint64 `json:"used"`
	Free       uint64 `json:"free"`
	FreeInodes uint64 `json:"freeInodes"`
}

type CPUStat struct {
	Id        string  `json:"id"`
	User      float64 `json:"user"`
	Nice      float64 `json:"nice"`
	System    float64 `json:"system"`
	Idle      float64 `json:"idle"`
	IOWait    uint64  `json:"iowait"`
	IRQ       uint64  `json:"irq"`
	SoftIRQ   uint64  `json:"softirq"`
	Steal     uint64  `json:"steal"`
	Guest     uint64  `json:"guest"`
	GuestNice uint64  `json:"guest_nice"`
}

type LoadAvg struct {
	Last1Min       float64 `json:"last1min"`
	Last5Min       float64 `json:"last5min"`
	Last15Min      float64 `json:"last15min"`
	ProcessRunning int     `json:"process_running"`
	ProcessTotal   int     `json:"process_total"`
	ProcessCreated int     `json:"process_created"`
	ProcessBlocked int     `json:"process_blocked"`
}

func ReadMemInfo() (*MemInfo, error) {
	v, err := mem.VirtualMemory()

	if err != nil {
		return nil, err
	}

	return &MemInfo{
		MemTotal:     v.Total,
		MemFree:      v.Free,
		MemAvailable: v.Available,
		Buffers:      v.Buffers,
		Cached:       v.Cached,
		SwapCached:   v.SwapCached,
		Active:       v.Active,
		Inactive:     v.Inactive,
		SwapTotal:    v.SwapTotal,
		SwapFree:     v.SwapFree,
		Shmem:        v.Shared,
		VmallocTotal: v.VmallocTotal,
		VmallocUsed:  v.VmallocUsed,
		VmallocChunk: v.VmallocChunk,
	}, nil
}

func ReadDisk(path string) (*Disk, error) {
	d, err := disk.Usage(path)

	if err != nil {
		return nil, err
	}

	return &Disk{
		All:        d.Total,
		Used:       d.Used,
		Free:       d.Free,
		FreeInodes: d.InodesFree,
	}, nil
}

func ReadStat() (*CPUStat, error) {
	c, err := cpu.Times(false)

	if err != nil {
		return nil, err
	}

	return &CPUStat{
		Id:     "cpu-all",
		User:   c[0].User,
		Nice:   c[0].Nice,
		System: c[0].System,
		Idle:   c[0].Idle,
	}, nil
}

func CpuUsage() (float64, error) {
	c, err := cpu.Percent(14500*time.Millisecond, false)

	if err != nil {
		return 0, err
	}

	return c[0], nil
}

func ReadLoadAvg() (*LoadAvg, error) {
	l, err := load.Avg()

	if err != nil {
		return nil, err
	}

	lm, err := load.Misc()

	if err != nil {
		return nil, err
	}

	return &LoadAvg{
		Last1Min:       l.Load1,
		Last5Min:       l.Load5,
		Last15Min:      l.Load15,
		ProcessRunning: lm.ProcsRunning,
		ProcessCreated: lm.ProcsCreated,
		ProcessBlocked: lm.ProcsBlocked,
		ProcessTotal:   lm.ProcsTotal,
	}, nil
}
