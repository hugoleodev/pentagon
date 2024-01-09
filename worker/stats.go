package worker

import (
	"log"

	"github.com/hugoleodev/pentagon/stats"
)

type Stats struct {
	Memory    *stats.MemInfo
	Disk      *stats.Disk
	Cpu       *stats.CPUStat
	Load      *stats.LoadAvg
	TaskCount int
}

func (s *Stats) MemTotalKb() uint64 {
	return s.Memory.MemTotal
}

func (s *Stats) MemAvailableKb() uint64 {
	return s.Memory.MemAvailable
}

func (s *Stats) MemUsedKb() uint64 {
	return s.Memory.MemTotal - s.Memory.MemAvailable
}

func (s *Stats) MemUsedPercent() uint64 {
	return s.Memory.MemAvailable / s.Memory.MemTotal
}

func (s *Stats) DiskTotal() uint64 {
	return s.Disk.All
}

func (s *Stats) DiskFree() uint64 {
	return s.Disk.Free
}

func (s *Stats) DiskUsed() uint64 {
	return s.Disk.Used
}

func (s *Stats) CpuUsage() float64 {
	u, err := stats.CpuUsage()

	if err != nil {
		log.Printf("Error reading cpu usage: %v\n", err)
		return 0
	}

	return u
}

func GetStats() *Stats {
	return &Stats{
		Memory: GetMemoryInfo(),
		Disk:   GetDiskInfo(),
		Cpu:    GetCpuStats(),
		Load:   GetLoadAvg(),
	}
}

func GetMemoryInfo() *stats.MemInfo {
	memstats, err := stats.ReadMemInfo()

	if err != nil {
		log.Printf("Error reading meminfo: %v\n", err)
		return &stats.MemInfo{}
	}

	return memstats
}

func GetDiskInfo() *stats.Disk {
	diskstats, err := stats.ReadDisk("/")
	if err != nil {
		log.Printf("Error reading from /: %v\n", err)
		return &stats.Disk{}
	}

	return diskstats
}

func GetCpuStats() *stats.CPUStat {
	s, err := stats.ReadStat()
	if err != nil {
		log.Printf("Error reading cpu stats: %v\n", err)
		return &stats.CPUStat{}
	}

	return s
}

func GetLoadAvg() *stats.LoadAvg {
	s, err := stats.ReadLoadAvg()
	if err != nil {
		log.Printf("Error reading loadavg: %v\n", err)
		return &stats.LoadAvg{}
	}

	return s
}
