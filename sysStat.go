package main

import (
	"bytes"
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

func Stat(t string) string {
	checkErr := func(err error) string {
		return "系统酱正在食用作死药丸中..."
	}
	switch t {
	case "free":
		m, err := mem.VirtualMemory()
		checkErr(err)
		s, err := mem.SwapMemory()
		checkErr(err)
		mem := new(runtime.MemStats)
		runtime.ReadMemStats(mem)
		return fmt.Sprintf(
			"全局:\n"+
				"Total: %s Free: %s\nUsed: %s %s%%\nCache: %s\n"+
				"Swap:\nTotal: %s Free: %s\n Used: %s %s%%\n"+
				"群组娘:\n"+
				"Allocated: %s\nTotal Allocated: %s\nSystem: %s\n",
			humanByte(m.Total, m.Free, m.Used, m.UsedPercent, m.Cached,
				s.Total, s.Free, s.Used, s.UsedPercent,
				mem.Alloc, mem.TotalAlloc, mem.Sys)...,
		)
	case "df":
		fs, err := disk.DiskPartitions(false)
		checkErr(err)
		var buf bytes.Buffer
		for k := range fs {
			du, err := disk.DiskUsage(fs[k].Mountpoint)
			switch {
			case err != nil, du.UsedPercent == 0, du.Free == 0:
				continue
			}
			f := fmt.Sprintf("Mountpoint: %s Type: %s \n"+
				"Total: %s Free: %s \nUsed: %s %s%%\n",
				humanByte(fs[k].Mountpoint, fs[k].Fstype,
					du.Total, du.Free, du.Used, du.UsedPercent)...,
			)
			buf.WriteString(f)
		}
		return buf.String()
	case "os":
		h, err := host.HostInfo()
		checkErr(err)
		l, err := load.LoadAvg()
		checkErr(err)
		c, err := cpu.CPUPercent(time.Second*3, false)
		checkErr(err)
		return fmt.Sprintf(
			"OSRelease: %s\nHostName: %s\nLoadAdv: %.2f %.2f %.2f\n"+
				"Goroutine: %d\nCPU: %.2f%%",
			h.Platform, h.Hostname, l.Load1, l.Load5, l.Load15,
			runtime.NumGoroutine(), c[0],
		)
	default:
		return "欢迎来到未知领域(ゝ∀･)"
	}
}
