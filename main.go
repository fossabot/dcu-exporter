package main

import (
	"bytes"
	"dcu-exporter-v2/pkg/podresources"
	"flag"
	"fmt"
	"github.com/Project-HAMi/dcu-dcgm/pkg/dcgm"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
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
		"hygon.com/dcu-share",
		"hygon.com/dcunum", // 4NF HAMi   Cannot prompt because user interactivity has been disabled
	}
	maxSize = 1024 * 1024 * 16 // 16 Mb

	portFlag = flag.Int("port", 16080, "Port number for the exporter")
)

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
	dcuComputeUnitCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_compute_unit_count",
			Help: "dcu metrics of gauge",
		},
		[]string{"device_id", "minor_number", "name", "node", "pcieBus_number"},
	)
	dcuComputeUnitRemainingCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_compute_unit_remaining_count",
			Help: "dcu metrics of gauge",
		},
		[]string{"device_id", "minor_number", "name", "node", "pcieBus_number"},
	)
	dcuMemoryRemaining = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_memory_remaining",
			Help: "dcu metrics of gauge",
		},
		[]string{"device_id", "minor_number", "name", "node", "pcieBus_number"},
	)

	// 虚拟设备collector
	vdcuComputeUnitCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "vdcu_compute_unit_count",
			Help: "vdcu metrics of gauge",
		},
		[]string{"device_id", "minor_number", "name", "node", "pcieBus_number", "container_id"},
	)
	vdcuGlobalMemSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "vdcu_global_memory_size",
			Help: "vdcu metrics of gauge",
		},
		[]string{"device_id", "minor_number", "name", "node", "pcieBus_number", "container_id"},
	)
	vdcuUsageMemSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "vdcu_usage_memory_size",
			Help: "vdcu metrics of gauge",
		},
		[]string{"device_id", "minor_number", "name", "node", "pcieBus_number", "container_id"},
	)
	vdcuPercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "vdcu_percent",
			Help: "vdcu metrics of gauge",
		},
		[]string{"device_id", "minor_number", "name", "node", "pcieBus_number", "container_id"},
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
	dcuComputeUnitCount.Reset()
	dcuComputeUnitRemainingCount.Reset()
	dcuMemoryRemaining.Reset()
	vdcuCollectorReset()
}

func vdcuCollectorReset() {
	vdcuComputeUnitCount.Reset()
	vdcuGlobalMemSize.Reset()
	vdcuUsageMemSize.Reset()
	vdcuPercent.Reset()
}

// 采集数据并设置collector值
func recordMetrics() {
	go func() {
		for {
			deviceInfos, err := dcgm.AllDeviceInfos()
			if err != nil {
				glog.Errorf("Get device metrics error: %v ", err)
				time.Sleep(10 * time.Second) // 等待一段时间后重试
				continue
			}
			fmt.Printf("Get devices number : %v \n", len(deviceInfos))
			// 如果出现device数量变化了，就要重置collector
			if deviceNumTag != len(deviceInfos) {
				deviceNumTag = len(deviceInfos)
				collectorReset()
			}
			if len(deviceInfos) > 0 {
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
				vdeviceNumTag := make(map[string]int)
				for _, info := range deviceInfos {
					dcuTemp.With(prometheus.Labels{
						"device_id":      info.Device.DeviceId,
						"minor_number":   strconv.Itoa(info.Device.MinorNumber),
						"name":           info.Device.SubSystemName,
						"node":           nodeName,
						"pcieBus_number": info.Device.PciBusNumber,
					}).Set(info.Device.Temperature)
					dcuPowerUsage.With(prometheus.Labels{
						"device_id":      info.Device.DeviceId,
						"minor_number":   strconv.Itoa(info.Device.MinorNumber),
						"name":           info.Device.SubSystemName,
						"node":           nodeName,
						"pcieBus_number": info.Device.PciBusNumber,
					}).Set(info.Device.PowerUsage)
					dcuPowerCap.With(prometheus.Labels{
						"device_id":      info.Device.DeviceId,
						"minor_number":   strconv.Itoa(info.Device.MinorNumber),
						"name":           info.Device.SubSystemName,
						"node":           nodeName,
						"pcieBus_number": info.Device.PciBusNumber,
					}).Set(info.Device.PowerCap)
					dcuMemoryCapBytes.With(prometheus.Labels{
						"device_id":      info.Device.DeviceId,
						"minor_number":   strconv.Itoa(info.Device.MinorNumber),
						"name":           info.Device.SubSystemName,
						"node":           nodeName,
						"pcieBus_number": info.Device.PciBusNumber,
					}).Set(info.Device.MemoryCap)
					dcuUsedMemoryBytes.With(prometheus.Labels{
						"device_id":      info.Device.DeviceId,
						"minor_number":   strconv.Itoa(info.Device.MinorNumber),
						"name":           info.Device.SubSystemName,
						"node":           nodeName,
						"pcieBus_number": info.Device.PciBusNumber,
					}).Set(info.Device.MemoryUsed)
					dcuUtilizationRate.With(prometheus.Labels{
						"device_id":      info.Device.DeviceId,
						"minor_number":   strconv.Itoa(info.Device.MinorNumber),
						"name":           info.Device.SubSystemName,
						"node":           nodeName,
						"pcieBus_number": info.Device.PciBusNumber,
					}).Set(info.Device.UtilizationRate)
					dcuPcieBwMb.With(prometheus.Labels{
						"device_id":      info.Device.DeviceId,
						"minor_number":   strconv.Itoa(info.Device.MinorNumber),
						"name":           info.Device.SubSystemName,
						"node":           nodeName,
						"pcieBus_number": info.Device.PciBusNumber,
					}).Set(info.Device.PcieBwMb)
					dcuSclk.With(prometheus.Labels{
						"device_id":      info.Device.DeviceId,
						"minor_number":   strconv.Itoa(info.Device.MinorNumber),
						"name":           info.Device.SubSystemName,
						"node":           nodeName,
						"pcieBus_number": info.Device.PciBusNumber,
					}).Set(info.Device.Clk)
					dcuComputeUnitCount.With(prometheus.Labels{
						"device_id":      info.Device.DeviceId,
						"minor_number":   strconv.Itoa(info.Device.MinorNumber),
						"name":           info.Device.SubSystemName,
						"node":           nodeName,
						"pcieBus_number": info.Device.PciBusNumber,
					}).Set(info.Device.ComputeUnitCount)
					dcuComputeUnitRemainingCount.With(prometheus.Labels{
						"device_id":      info.Device.DeviceId,
						"minor_number":   strconv.Itoa(info.Device.MinorNumber),
						"name":           info.Device.SubSystemName,
						"node":           nodeName,
						"pcieBus_number": info.Device.PciBusNumber,
					}).Set(float64(info.Device.ComputeUnitRemainingCount))
					dcuMemoryRemaining.With(prometheus.Labels{
						"device_id":      info.Device.DeviceId,
						"minor_number":   strconv.Itoa(info.Device.MinorNumber),
						"name":           info.Device.SubSystemName,
						"node":           nodeName,
						"pcieBus_number": info.Device.PciBusNumber,
					}).Set(float64(info.Device.MemoryRemaining))

					// vdcu metrics
					fmt.Printf("vdeviceNumTag[info.Device.PciBusNumber] : %v \n", vdeviceNumTag[info.Device.PciBusNumber])
					fmt.Printf("len(info.VirtualDevices) : %v \n", len(info.VirtualDevices))

					if vdeviceNumTag[info.Device.PciBusNumber] != len(info.VirtualDevices) {
						fmt.Printf("info.VirtualDevices : %v \n", info.VirtualDevices)
						vdeviceNumTag[info.Device.PciBusNumber] = len(info.VirtualDevices)
						fmt.Printf("vdeviceNumTag[info.Device.PciBusNumber] : %v \n", vdeviceNumTag[info.Device.PciBusNumber])
						vdcuCollectorReset()
					}
					for _, virtualDevice := range info.VirtualDevices {
						vdcuComputeUnitCount.With(prometheus.Labels{
							"device_id":      strconv.Itoa(virtualDevice.DeviceID),
							"minor_number":   strconv.Itoa(virtualDevice.VMinorNumber),
							"node":           nodeName,
							"name":           info.Device.SubSystemName,
							"pcieBus_number": info.Device.PciBusNumber,
							"container_id":   strconv.FormatUint(virtualDevice.ContainerID, 10),
						}).Set(float64(virtualDevice.ComputeUnitCount))
						vdcuGlobalMemSize.With(prometheus.Labels{
							"device_id":      strconv.Itoa(virtualDevice.DeviceID),
							"minor_number":   strconv.Itoa(virtualDevice.VMinorNumber),
							"node":           nodeName,
							"name":           info.Device.SubSystemName,
							"pcieBus_number": info.Device.PciBusNumber,
							"container_id":   strconv.FormatUint(virtualDevice.ContainerID, 10),
						}).Set(float64(virtualDevice.GlobalMemSize))
						vdcuUsageMemSize.With(prometheus.Labels{
							"device_id":      strconv.Itoa(virtualDevice.DeviceID),
							"minor_number":   strconv.Itoa(virtualDevice.VMinorNumber),
							"node":           nodeName,
							"name":           info.Device.SubSystemName,
							"pcieBus_number": info.Device.PciBusNumber,
							"container_id":   strconv.FormatUint(virtualDevice.ContainerID, 10),
						}).Set(float64(virtualDevice.UsageMemSize))
						vdcuPercent.With(prometheus.Labels{
							"device_id":      strconv.Itoa(virtualDevice.DeviceID),
							"minor_number":   strconv.Itoa(virtualDevice.VMinorNumber),
							"node":           nodeName,
							"name":           info.Device.SubSystemName,
							"pcieBus_number": info.Device.PciBusNumber,
							"container_id":   strconv.FormatUint(virtualDevice.ContainerID, 10),
						}).Set(float64(virtualDevice.Percent))
					}

					deviceIDs[info.Device.PciBusNumber] = info.Device.DeviceId
					deviceMinors[info.Device.PciBusNumber] = strconv.Itoa(info.Device.MinorNumber)
					deviceName[info.Device.PciBusNumber] = info.Device.SubSystemName
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
			time.Sleep(10 * time.Second)
		}
	}()
}

func main() {
	glog.Infof("🚀 🚀 🚀  DCU exporter start ...")

	fmt.Printf("Init ROCm smi: %v \n", dcgm.Init())
	defer func() {
		err := dcgm.ShutDown()
		if err != nil {
			glog.Errorf("DCU exporter shutdown Error: %v ", err)
			return
		}
	}()

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
	registry.MustRegister(dcuComputeUnitCount)
	registry.MustRegister(dcuComputeUnitRemainingCount)
	registry.MustRegister(dcuMemoryRemaining)
	registry.MustRegister(vdcuComputeUnitCount)
	registry.MustRegister(vdcuGlobalMemSize)
	registry.MustRegister(vdcuUsageMemSize)
	registry.MustRegister(vdcuPercent)

	recordMetrics()

	flag.Parse()
	port := fmt.Sprintf("%d", *portFlag)
	glog.Infof("🚀 🚀 🚀  DCU exporter start on port %d ...", *portFlag)
	if port == "16080" {
		port = os.Getenv("DCU_EXPORTER_LISTEN")
		if port == "" {
			port = "16080"
		}
	}
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry}))
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
