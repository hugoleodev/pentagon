package worker

import (
	"log"

	"github.com/c9s/goprocinfo/linux"
)

type Stats struct {
	Memory *linux.MemInfo
	Disk   *linux.Disk
	Cpu    *linux.CPUStat
	Load   *linux.LoadAvg
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
	idle := s.Cpu.Idle + s.Cpu.IOWait
	nonIdle := s.Cpu.User + s.Cpu.Nice + s.Cpu.System + s.Cpu.IRQ + s.Cpu.SoftIRQ + s.Cpu.Steal
	total := idle + nonIdle

	if total == 0 {
		return 0
	}

	return (float64(total) - float64(idle)) / float64(total)
}

func GetStats() *Stats {
	return &Stats{
		Memory: GetMemoryInfo(),
		Disk:   GetDiskInfo(),
		Cpu:    GetCpuStats(),
		Load:   GetLoadAvg(),
	}
}

func GetMemoryInfo() *linux.MemInfo {
	memstats, err := linux.ReadMemInfo("/proc/meminfo")

	if err != nil {
		log.Printf("Error reading /proc/meminfo: %v\n", err)
		return &linux.MemInfo{}
	}

	return memstats
}

func GetDiskInfo() *linux.Disk {
	diskstats, err := linux.ReadDisk("/")
	if err != nil {
		log.Printf("Error reading from /: %v\n", err)
		return &linux.Disk{}
	}

	return diskstats
}

func GetCpuStats() *linux.CPUStat {
	stats, err := linux.ReadStat("/proc/stat")
	if err != nil {
		log.Printf("Error reading /proc/stat: %v\n", err)
		return &linux.CPUStat{}
	}

	return &stats.CPUStatAll
}

func GetLoadAvg() *linux.LoadAvg {
	stats, err := linux.ReadLoadAvg("/proc/loadavg")
	if err != nil {
		log.Printf("Error reading /proc/loadavg: %v\n", err)
		return &linux.LoadAvg{}
	}

	return stats
}
