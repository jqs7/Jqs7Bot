package plugin

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jqs7/Jqs7Bot/conf"
	"github.com/jqs7/bb"
	"github.com/pyk/byten"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

type Reload struct{ bb.Base }

func (r *Reload) Run() {
	conf.LoadConf()
	r.NewMessage(r.ChatID,
		"群组娘已完成弹药重装(ゝ∀･)").Send()
}

type Stat struct{ Default }

func (s *Stat) Run() {
	if s.isMaster() {
		command := strings.TrimLeft(s.Args[0], "/")
		s.NewMessage(s.ChatID, s.stat(command)).Send()
	}
}

func (st *Stat) stat(t string) string {
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
			st.humanByte(m.Total, m.Free, m.Used, m.UsedPercent, m.Cached,
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
				st.humanByte(fs[k].Mountpoint, fs[k].Fstype,
					du.Total, du.Free, du.Used, du.UsedPercent)...,
			)
			buf.WriteString(f)
		}
		return buf.String()
	case "os":
		h, err := host.HostInfo()
		checkErr(err)
		uptime := time.Duration(time.Now().Unix()-int64(h.Uptime)) * time.Second
		l, err := load.LoadAvg()
		checkErr(err)
		c, err := cpu.CPUPercent(time.Second*3, false)
		checkErr(err)
		return fmt.Sprintf(
			"OSRelease: %s\nHostName: %s\nUptime: %s\nLoadAvg: %.2f %.2f %.2f\n"+
				"Goroutine: %d\nCPU: %.2f%%",
			h.Platform, h.Hostname, uptime.String(), l.Load1, l.Load5, l.Load15,
			runtime.NumGoroutine(), c[0],
		)
	case "redis":
		info := conf.Redis.Info().Val()
		if info != "" {
			infos := strings.Split(info, "\r\n")
			infoMap := make(map[string]string)
			for k := range infos {
				line := strings.Split(infos[k], ":")
				if len(line) > 1 {
					infoMap[line[0]] = line[1]
				}
			}
			DBSize := conf.Redis.DbSize().Val()
			return fmt.Sprintf("Redis Version: %s\nOS: %s\nUsed Memory: %s\n"+
				"Used Memory Peak: %s\nDB Size: %d\n",
				infoMap["redis_version"], infoMap["os"], infoMap["used_memory_human"],
				infoMap["used_memory_peak_human"], DBSize)
		}
		return ""
	default:
		return "欢迎来到未知领域(ゝ∀･)"
	}
}

func (s *Stat) humanByte(in ...interface{}) (out []interface{}) {
	for _, v := range in {
		switch v.(type) {
		case int, uint64:
			s := fmt.Sprintf("%d", v)
			i, _ := strconv.ParseInt(s, 10, 64)
			out = append(out, byten.Size(i))
		case float64:
			s := fmt.Sprintf("%.2f", v)
			out = append(out, s)
		default:
			out = append(out, v)
		}
	}
	return out
}
