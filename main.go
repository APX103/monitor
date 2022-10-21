package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/v3/cpu"
)

type InfluxClient struct {
	bucket string
	org    string
	token  string
	url    string
}

type SysInfo struct {
	CpuCount   int
	CpuPercent float64
	MemTotal   uint64
	MemFree    uint64
	MemPercent float64
}

type GpuInfo struct {
	GpuUtilization    int64
	GpuMemUtilization int64
	GpuMemTotal       int64
	GpuMemUsed        int64
	GpuMemFree        int64
}

func Str2Num(s string) int64 {
	pattern := regexp.MustCompile(`(\d+)`)
	numberStrings := pattern.FindAllStringSubmatch(strings.Split(s, ", ")[0], -1)
	numbers := make([]int64, len(numberStrings))
	for i, numberString := range numberStrings {
		number, err := strconv.Atoi(numberString[1])
		if err != nil {
			panic(err)
		}
		numbers[i] = int64(number)
	}
	return numbers[0]
}

func GetSysInfo() SysInfo {
	v, _ := mem.VirtualMemory()
	counts, _ := cpu.Counts(true)
	percent, _ := cpu.Percent(time.Second, false)
	return SysInfo{
		CpuCount:   counts,
		CpuPercent: percent[0],
		MemTotal:   v.Total,
		MemFree:    v.Free,
		MemPercent: v.UsedPercent,
	}
}

func GetGpuInfo() GpuInfo {
	cmd := exec.Command("nvidia-smi", "--query-gpu=memory.total,memory.free,memory.used,utilization.gpu,utilization.memory", "--format=csv")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error:can not obtain stdout pipe for command:%s\nmaybe because there is no nvidia-smi", err)
		panic("Error")
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("Error:The command is err,", err)
		panic("The command is err")
	}

	outputBuf := bufio.NewReader(stdout)
	outputBuf.ReadLine()
	output, _, err := outputBuf.ReadLine()
	if err != nil {
		if err.Error() != "EOF" {
			fmt.Printf("Error :%s\n", err)
		}
		return GpuInfo{
			GpuUtilization:    0,
			GpuMemUtilization: 0,
			GpuMemTotal:       0,
			GpuMemUsed:        0,
			GpuMemFree:        0,
		}
	}
	r := strings.Split(string(output), ", ")
	return GpuInfo{
		GpuUtilization:    Str2Num(r[3]),
		GpuMemUtilization: Str2Num(r[4]),
		GpuMemTotal:       Str2Num(r[0]),
		GpuMemUsed:        Str2Num(r[2]),
		GpuMemFree:        Str2Num(r[1]),
	}
}

func UpdateToInflux(p *InfluxClient) {
	client := influxdb2.NewClient(p.url, p.token)
	defer client.Close()
	writeAPI := client.WriteAPIBlocking(p.org, p.bucket)

	for {
		sysinfo := GetSysInfo()
		gpuinfo := GetGpuInfo()
		q := influxdb2.NewPoint("server_stat",
			map[string]string{"unit": "utilization"},
			map[string]interface{}{
				"CpuCount":          sysinfo.CpuCount,
				"CpuPercent":        sysinfo.CpuPercent,
				"MemTotal":          sysinfo.MemTotal,
				"MemFree":           sysinfo.MemFree,
				"MemPercent":        sysinfo.MemPercent,
				"GpuUtilization":    gpuinfo.GpuUtilization,
				"GpuMemUtilization": gpuinfo.GpuMemUtilization,
				"GpuMemTotal":       gpuinfo.GpuMemTotal,
				"GpuMemUsed":        gpuinfo.GpuMemUsed,
				"GpuMemFree":        gpuinfo.GpuMemFree},
			time.Now())
		writeAPI.WritePoint(context.Background(), q)
		time.Sleep(time.Second * 4)
		// fmt.Println("It's running")
	}

}

// influx setup --username lijialun --password pjlab666 --org pjlab --bucket monitor
func main() {
	bucket := flag.String("bucket", "monitor", "Name of bucket of influxdb2")
	org := flag.String("org", "pjlab", "Name of org of influxdb2")
	token := flag.String("token", "wnk9mEaeN6GH4ky3E6T48ED8qwqRq3sn0PsoycRxY01gLOqsEhLc3y2z0INJjHsumIyWfZ3kNann6YnZJRqpPA==", "API token of influxdb2")
	url := flag.String("url", "http://localhost:8086", "Url of influxdb2")

	flag.Parse()
	client := InfluxClient{
		bucket: *bucket,
		org:    *org,
		token:  *token,
		url:    *url,
	}
	UpdateToInflux(&client)
}
