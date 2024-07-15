package main

import (
	"bytes"
	"dcu-exporter-v2/pkg/podresources"
	"dcu-exporter-v2/pkg/shim"
	"fmt"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	timeout   = 10 * time.Second
	socket    = "/var/lib/kubelet/pod-resources/kubelet.sock"
	resources = []string{
		"hygon.com/dcu",
	}
	maxSize = 1024 * 1024 * 16 // 16 Mb
)

var type2name = map[string]string{
	"66a1": "WK100",
	"51b7": "Z100L",
	"52b7": "Z100L",
	"53b7": "Z100L",
	"54b7": "Z100L",
	"55b7": "Z100L",
	"56b7": "Z100L",
	"57b7": "Z100L",
	"61b7": "K100",
	"62b7": "K100",
	"63b7": "K100",
	"64b7": "K100",
	"65b7": "K100",
	"66b7": "K100",
	"67b7": "K100",
	"6210": "K100 AI",
	"6211": "K100 AI Liquid",
	"6212": "K100 AI Liquid",
}

// 定义collector
var (
	dcuTemp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_temp",
			Help: "dcu metrics of gauge",
		},
		[]string{"device_id", "minor_number", "name", "node", "pcieBus_number"},
	)
	dcuPowerUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_power_usage",
			Help: "dcu metrics of gauge",
		},
		[]string{"device_id", "minor_number", "name", "node", "pcieBus_number"},
	)

	dcuPowerCap = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_powercap",
			Help: "dcu metrics of gauge",
		},
		[]string{"device_id", "minor_number", "name", "node", "pcieBus_number"},
	)

	dcuSclk = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_sclk",
			Help: "dcu metrics of gauge",
		},
		[]string{"device_id", "minor_number", "name", "node", "pcieBus_number"},
	)

	dcuUtilizationRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_utilizationrate",
			Help: "dcu metrics of gauge",
		},
		[]string{"device_id", "minor_number", "name", "node", "pcieBus_number"},
	)

	dcuUsedMemoryBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_usedmemory_bytes",
			Help: "dcu metrics of gauge",
		},
		[]string{"device_id", "minor_number", "name", "node", "pcieBus_number"},
	)

	dcuMemoryCapBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_memorycap_bytes",
			Help: "dcu metrics of gauge",
		},
		[]string{"device_id", "minor_number", "name", "node", "pcieBus_number"},
	)

	dcuPcieBwMb = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_pciebw_mb",
			Help: "dcu metrics of gauge",
		},
		[]string{"device_id", "minor_number", "name", "node", "pcieBus_number"},
	)
	dcuContainer = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_containers",
			Help: "containers in dcu node",
		},
		[]string{"node", "minor_number", "pcieBus_number", "device_id", "name", "container", "dcu_pod_name", "dcu_pod_namespace"},
	)
)

var deviceNumTag = 0
var PodNumTag = 0

func collectorReset() {
	dcuTemp.Reset()
	dcuPowerUsage.Reset()
	dcuPowerCap.Reset()
	dcuSclk.Reset()
	dcuUtilizationRate.Reset()
	dcuUsedMemoryBytes.Reset()
	dcuMemoryCapBytes.Reset()
	dcuPcieBwMb.Reset()
	dcuContainer.Reset()
}

// 采集数据并设置collector值
func recordMetrics() {
	go func() {
		for {
			numMonitorDevices := shim.GO_rsmi_num_monitor_devices()
			fmt.Printf("Get devices number : %v \n", numMonitorDevices)
			// 如果出现device数量变化了，就要重置collector
			if deviceNumTag != numMonitorDevices {
				deviceNumTag = numMonitorDevices
				collectorReset()
			}
			if numMonitorDevices > 0 { // 存在 dcu 设备
				cmd := exec.Command("cat", "/etc/hostname")
				var out bytes.Buffer
				var stderr bytes.Buffer
				cmd.Stdout = &out
				cmd.Stderr = &stderr
				err := cmd.Run()
				if err != nil {
					fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
					return
				}
				nodeName := strings.TrimSpace(out.String())
				deviceIDs := make(map[string]string)
				deviceMinors := make(map[string]string)
				deviceName := make(map[string]string)
				for i := 0; i < numMonitorDevices; i++ {
					bdfid := shim.GO_rsmi_dev_pci_id_get(i)
					// 解析BDFID
					domain := (bdfid >> 32) & 0xffffffff
					bus := (bdfid >> 8) & 0xff
					dev := (bdfid >> 3) & 0x1f
					function := bdfid & 0x7
					// 格式化PCI ID
					picBusNumber := fmt.Sprintf("%04x:%02x:%02x.%x", domain, bus, dev, function)
					deviceId := shim.GO_rsmi_dev_serial_number_get(i)

					devId := shim.GO_rsmi_dev_id_get(i)
					subSystemName := type2name[fmt.Sprintf("%X", devId)]
					temperature := shim.GO_rsmi_dev_temp_metric_get(i, 0, shim.RSMI_TEMP_CURRENT)
					t, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(temperature)/1000.0), 64)
					fmt.Printf("🌡️  DCU[%v] temperature : %v \n", i, t)
					dcuTemp.With(prometheus.Labels{
						"device_id":      deviceId,
						"minor_number":   strconv.Itoa(i),
						"name":           subSystemName,
						"node":           nodeName,
						"pcieBus_number": picBusNumber,
					}).Set(t)

					powerUsage := shim.GO_rsmi_dev_power_ave_get(i, 0)
					pu, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(powerUsage)/1000000.0), 64)
					fmt.Printf("🔋 DCU[%v] power cap : %v \n", i, pu)
					dcuPowerUsage.With(prometheus.Labels{
						"device_id":      deviceId,
						"minor_number":   strconv.Itoa(i),
						"name":           subSystemName,
						"node":           nodeName,
						"pcieBus_number": picBusNumber,
					}).Set(pu)

					powerCap := shim.GO_rsmi_dev_power_cap_get(i, 0)
					pc, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(powerCap)/1000000.0), 64)
					fmt.Printf("\U0001FAAB DCU[%v] power usage : %v \n", i, pc)
					dcuPowerCap.With(prometheus.Labels{
						"device_id":      deviceId,
						"minor_number":   strconv.Itoa(i),
						"name":           subSystemName,
						"node":           nodeName,
						"pcieBus_number": picBusNumber,
					}).Set(pc)

					memoryCap := shim.GO_rsmi_dev_memory_total_get(i, shim.RSMI_MEM_TYPE_FIRST)
					mc, _ := strconv.ParseFloat(fmt.Sprintf("%f", float64(memoryCap)/1.0), 64)
					fmt.Printf(" DCU[%v] memory cap : %v \n", i, mc)
					dcuMemoryCapBytes.With(prometheus.Labels{
						"device_id":      deviceId,
						"minor_number":   strconv.Itoa(i),
						"name":           subSystemName,
						"node":           nodeName,
						"pcieBus_number": picBusNumber,
					}).Set(mc)

					memoryUsed := shim.GO_rsmi_dev_memory_usage_get(i, shim.RSMI_MEM_TYPE_FIRST)
					mu, _ := strconv.ParseFloat(fmt.Sprintf("%f", float64(memoryUsed)/1.0), 64)
					fmt.Printf(" DCU[%v] memory used : %v \n", i, mu)
					dcuUsedMemoryBytes.With(prometheus.Labels{
						"device_id":      deviceId,
						"minor_number":   strconv.Itoa(i),
						"name":           subSystemName,
						"node":           nodeName,
						"pcieBus_number": picBusNumber,
					}).Set(mu)

					utilizationRate := shim.GO_rsmi_dev_busy_percent_get(i)
					ur, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(utilizationRate)/1.0), 64)
					fmt.Printf(" DCU[%v] utilization rate : %v \n", i, ur)
					dcuUtilizationRate.With(prometheus.Labels{
						"device_id":      deviceId,
						"minor_number":   strconv.Itoa(i),
						"name":           subSystemName,
						"node":           nodeName,
						"pcieBus_number": picBusNumber,
					}).Set(ur)

					sent, received, maxPktSz := shim.GO_rsmi_dev_pci_throughput_get(i)
					pcieBwMb, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", float64(received+sent)*float64(maxPktSz)/1024.0/1024.0), 64)
					fmt.Printf(" DCU[%v] PCIE  bandwidth : %v \n", i, pcieBwMb)
					dcuPcieBwMb.With(prometheus.Labels{
						"device_id":      deviceId,
						"minor_number":   strconv.Itoa(i),
						"name":           subSystemName,
						"node":           nodeName,
						"pcieBus_number": picBusNumber,
					}).Set(pcieBwMb)

					clk := shim.GO_rsmi_dev_gpu_clk_freq_get(i, shim.RSMI_CLK_TYPE_SYS)
					sclk, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(clk.Frequency[clk.Current])/1000000.0), 64)
					fmt.Printf(" DCU[%v] SCLK : %v \n", i, sclk)
					dcuSclk.With(prometheus.Labels{
						"device_id":      deviceId,
						"minor_number":   strconv.Itoa(i),
						"name":           subSystemName,
						"node":           nodeName,
						"pcieBus_number": picBusNumber,
					}).Set(sclk)

					deviceIDs[picBusNumber] = deviceId
					deviceMinors[picBusNumber] = strconv.Itoa(i)
					deviceName[picBusNumber] = subSystemName
				}

				// 获取pod resources指标数据
				podresource := podresources.NewPodResourcesClient(timeout, socket, resources, maxSize)
				podInfoMap, err := podresource.GetDeviceToPodInfo()
				fmt.Printf(" podInfoMap: %v \n", podInfoMap)
				if PodNumTag != len(podInfoMap) {
					PodNumTag = len(podInfoMap)
					dcuContainer.Reset()
				}
				for id, podInfo := range podInfoMap {
					dcuContainer.With(prometheus.Labels{
						"node":              nodeName,
						"minor_number":      deviceMinors[id],
						"pcieBus_number":    id,
						"device_id":         deviceIDs[id],
						"name":              deviceName[id],
						"container":         podInfo.Container,
						"dcu_pod_name":      podInfo.Pod,
						"dcu_pod_namespace": podInfo.Namespace,
					}).Set(1)
				}
			}

			time.Sleep(5 * time.Second)
		}
	}()
}

func main() {
	glog.Infof("🚀 🚀 🚀  DCU exporter start ...")

	fmt.Printf("Init ROCm smi: %v \n", shim.GO_rsmi_init())
	defer shim.GO_rsmi_shutdown()

	// 这里用自定义注册表，可以使返回的数据比较简洁
	registry := prometheus.NewRegistry()
	// NewGoCollector和NewProcessCollector采集的内容是默认注册表自带的，需要时可以打开注释即可
	/*	registry.MustRegister(collectors.NewGoCollector())
		registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))*/

	// 注册采集器
	registry.MustRegister(dcuTemp)
	registry.MustRegister(dcuPowerUsage)
	registry.MustRegister(dcuPowerCap)
	registry.MustRegister(dcuUtilizationRate)
	registry.MustRegister(dcuUsedMemoryBytes)
	registry.MustRegister(dcuMemoryCapBytes)
	registry.MustRegister(dcuPcieBwMb)
	registry.MustRegister(dcuSclk)
	registry.MustRegister(dcuContainer)

	recordMetrics()
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry}))
	log.Fatal(http.ListenAndServe(":16081", nil))
}
