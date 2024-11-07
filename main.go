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
	timeout = 10 * time.Second
	socket  = "/var/lib/kubelet/pod-resources/kubelet.sock"

	resources = []string{
		"hygon.com/dcu",
	}

	vdcuResources = []string{
		"hygon.com/dcu-share",
	}

	maxSize = 1024 * 1024 * 16 // 16 Mb

	dcuLabels    = []string{"device_id", "minor_number", "name", "node", "pcieBus_number", "dcu_pod_namespace", "dcu_pod_name", "container"}
	vDcuLabels   = []string{"vdcu_minor_number", "vdcu_computer_unit", "vdcu_memory_cap", "device_id", "minor_number", "name", "node", "dcu_pod_namespace", "dcu_pod_name", "container"}
	dcuErrLabels = []string{"device_id", "minor_number", "name", "node", "pcieBus_number", "dcu_pod_namespace", "dcu_pod_name", "container", "block_type"}

	portFlag = flag.Int("port", 16080, "Port number for the exporter")
)

// 定义collector
var (
	dcuTemp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_temp",
			Help: "dcu metrics of gauge",
		},
		dcuLabels,
	)
	dcuPowerUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_power_usage",
			Help: "dcu metrics of gauge",
		},
		dcuLabels,
	)

	dcuPowerCap = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_powercap",
			Help: "dcu metrics of gauge",
		},
		dcuLabels,
	)

	dcuSclk = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_sclk",
			Help: "dcu metrics of gauge",
		},
		dcuLabels,
	)

	dcuUtilizationRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_utilizationrate",
			Help: "dcu metrics of gauge",
		},
		dcuLabels,
	)

	dcuUsedMemoryBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_usedmemory_bytes",
			Help: "dcu metrics of gauge",
		},
		dcuLabels,
	)

	dcuMemoryCapBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_memorycap_bytes",
			Help: "dcu metrics of gauge",
		},
		dcuLabels,
	)

	dcuPcieBwMb = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_pciebw_mb",
			Help: "dcu metrics of gauge",
		},
		dcuLabels,
	)

	dcuComputeUnitCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_compute_unit_count",
			Help: "dcu metrics of gauge",
		},
		dcuLabels,
	)
	dcuComputeUnitRemainingCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_compute_unit_remaining_count",
			Help: "dcu metrics of gauge",
		},
		dcuLabels,
	)
	dcuMemoryRemaining = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_memory_remaining",
			Help: "dcu metrics of gauge",
		},
		dcuLabels,
	)
	dcuCE = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_ce_count",
			Help: "dcu metrics of gauge",
		},
		dcuErrLabels,
	)
	dcuUE = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dcu_ue_count",
			Help: "dcu metrics of gauge",
		},
		dcuErrLabels,
	)
)

// 定义vdcu collector
var (
	vdcuTemp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "vdcu_temp",
			Help: "vdcu metrics of gauge",
		},
		vDcuLabels,
	)

	vdcuSclk = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "vdcu_sclk",
			Help: "vdcu metrics of gauge",
		},
		vDcuLabels,
	)

	vdcuUtilizationRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "vdcu_utilizationrate",
			Help: "vdcu metrics of gauge",
		},
		vDcuLabels,
	)

	vdcuUsedMemoryBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "vdcu_usedmemory_bytes",
			Help: "vdcu metrics of gauge",
		},
		vDcuLabels,
	)
)

var deviceNumTag = 0
var PodNumTag = 0
var podInfoMap = make(map[string]podresources.PodInfo)

var vdeviceNumTag = 0
var podDCUShareNumTag = 0
var podDCUDynamicNumTag = 0
var podDCUShareInfoMap = make(map[string]podresources.PodInfo)

func collectorReset() {
	dcuTemp.Reset()
	dcuPowerUsage.Reset()
	dcuPowerCap.Reset()
	dcuSclk.Reset()
	dcuUtilizationRate.Reset()
	dcuUsedMemoryBytes.Reset()
	dcuMemoryCapBytes.Reset()
	dcuPcieBwMb.Reset()
	dcuComputeUnitCount.Reset()
	dcuComputeUnitRemainingCount.Reset()
	dcuMemoryRemaining.Reset()
	dcuCE.Reset()
	dcuUE.Reset()
}

func vdcuCollectorReset() {
	vdcuTemp.Reset()
	vdcuSclk.Reset()
	vdcuUtilizationRate.Reset()
	vdcuUsedMemoryBytes.Reset()
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
			glog.Infof("Get devices number : %d \n", len(deviceInfos))

			cmd := exec.Command("cat", "/etc/hostname")
			var out bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				glog.Infof(fmt.Sprint(err) + ": " + stderr.String())
				return
			}
			nodeName := strings.TrimSpace(out.String())
			glog.Infof("Get NodeName : %s \n", nodeName)

			//获取虚拟设备
			var vdcuInfoMap = make(map[string]podresources.VitualDeviceInfo)
			for _, deviceInfo := range deviceInfos {
				for _, vdeviceInfo := range deviceInfo.VirtualDevices {
					var vdcuInfo = podresources.VitualDeviceInfo{
						VitualDeviceIndex: vdeviceInfo.VMinorNumber,
						DeviceIndex:       deviceInfo.Device.MinorNumber,
						DeviceID:          deviceInfo.Device.DeviceId,
						SubSystemName:     deviceInfo.Device.SubSystemName,
						Temperature:       deviceInfo.Device.Temperature,
						Clk:               deviceInfo.Device.Clk,
						ComputeUnitCount:  vdeviceInfo.ComputeUnitCount,
						MemoryCap:         int(vdeviceInfo.GlobalMemSize),
						MemoryUsed:        int(vdeviceInfo.UsageMemSize),
						UtilizationRate:   vdeviceInfo.Percent,
					}
					vdcuInfoMap[strconv.Itoa(vdeviceInfo.VMinorNumber)] = vdcuInfo
				}
			}
			glog.Infof("Get vdcu info : %v \n", vdcuInfoMap)

			// 获取pod resources指标数据
			if fileExists(socket) {
				glog.Infof("K8S exists~")
				podresource := podresources.NewPodResourcesClient(timeout, socket, resources, maxSize)
				podInfoMap, _ = podresource.GetDeviceToPodInfo()
				glog.Infof(" dcuPodInfoMap: %v \n", podInfoMap)

				podDCUShareResource := podresources.NewPodResourcesClient(timeout, socket, vdcuResources, maxSize)
				podDCUShareInfoMap, _ = podDCUShareResource.GetDeviceToPodInfo()
				glog.Infof(" podDCUShareInfoMap: %v \n", podDCUShareInfoMap)
			}

			// 如果出现device或Pod数量变化了，就要重置collector
			//if deviceNumTag != len(deviceInfos) || PodNumTag != len(podInfoMap) {
			//	deviceNumTag = len(deviceInfos)
			//	PodNumTag = len(podInfoMap)
			//	collectorReset()
			//}
			collectorReset()

			// 获取pod resources指标数据(dcunum)
			podDCUDynamicInfoMap, err := podresources.GetVDCUPodInfo()
			if err != nil {
				glog.Errorf("Get vdcu pod info error: %v ", err)
				continue
			}
			glog.Infof(" podDCUDynamicInfoMap: %v \n", podDCUDynamicInfoMap)

			//如果vdcu的Device或Pod数量变化了，就重置vdcu collector
			//if vdeviceNumTag != len(vdcuInfoMap) || podDCUShareNumTag != len(podDCUShareInfoMap) || podDCUDynamicNumTag != len(podDCUDynamicInfoMap) {
			//	vdeviceNumTag = len(vdcuInfoMap)
			//	podDCUShareNumTag = len(podDCUShareInfoMap)
			//	podDCUDynamicNumTag = len(podDCUDynamicInfoMap)
			//	vdcuCollectorReset()
			//}
			vdcuCollectorReset()

			if len(deviceInfos) > 0 {
				dcuPrometheusLabels := prometheus.Labels{}

				for _, info := range deviceInfos {
					podInfo, exists := podInfoMap[info.Device.PciBusNumber]
					if exists {
						dcuPrometheusLabels = prometheus.Labels{
							"device_id":         info.Device.DeviceId,
							"minor_number":      strconv.Itoa(info.Device.MinorNumber),
							"name":              info.Device.SubSystemName,
							"node":              nodeName,
							"pcieBus_number":    info.Device.PciBusNumber,
							"dcu_pod_namespace": podInfo.Namespace,
							"dcu_pod_name":      podInfo.Pod,
							"container":         podInfo.Container,
						}
					} else {
						dcuPrometheusLabels = prometheus.Labels{
							"device_id":         info.Device.DeviceId,
							"minor_number":      strconv.Itoa(info.Device.MinorNumber),
							"name":              info.Device.SubSystemName,
							"node":              nodeName,
							"pcieBus_number":    info.Device.PciBusNumber,
							"dcu_pod_namespace": "",
							"dcu_pod_name":      "",
							"container":         "",
						}
					}

					for _, errorInfo := range info.Device.BlocksInfos {
						dcuPrometheusLabels["block_type"] = errorInfo.Block
						dcuCE.With(dcuPrometheusLabels).Set(float64(errorInfo.CE))
						dcuUE.With(dcuPrometheusLabels).Set(float64(errorInfo.UE))
					}

					delete(dcuPrometheusLabels, "block_type")

					dcuTemp.With(dcuPrometheusLabels).Set(info.Device.Temperature)
					dcuPowerUsage.With(dcuPrometheusLabels).Set(info.Device.PowerUsage)
					dcuPowerCap.With(dcuPrometheusLabels).Set(info.Device.PowerCap)
					dcuMemoryCapBytes.With(dcuPrometheusLabels).Set(info.Device.MemoryCap)
					dcuUsedMemoryBytes.With(dcuPrometheusLabels).Set(info.Device.MemoryUsed)
					dcuUtilizationRate.With(dcuPrometheusLabels).Set(info.Device.UtilizationRate)
					dcuPcieBwMb.With(dcuPrometheusLabels).Set(info.Device.PcieBwMb)
					dcuSclk.With(dcuPrometheusLabels).Set(info.Device.Clk)
					dcuComputeUnitCount.With(dcuPrometheusLabels).Set(info.Device.ComputeUnitCount)
					dcuComputeUnitRemainingCount.With(dcuPrometheusLabels).Set(float64(info.Device.ComputeUnitRemainingCount))
					dcuMemoryRemaining.With(dcuPrometheusLabels).Set(float64(info.Device.MemoryRemaining))

					// dcu info
					glog.Infof("dcu info : %v \n", info)
				}
			}

			if len(vdcuInfoMap) > 0 {
				vdcuPrometheusLabels := prometheus.Labels{}

				for _, info := range vdcuInfoMap {
					podShareInfo, existsShare := podDCUShareInfoMap["vdev"+strconv.Itoa(info.VitualDeviceIndex)]
					podDynamicInfo, existsDyanamic := podDCUDynamicInfoMap[strconv.Itoa(info.VitualDeviceIndex)]
					if existsShare {
						vdcuPrometheusLabels = prometheus.Labels{
							"device_id":          info.DeviceID,
							"minor_number":       strconv.Itoa(info.DeviceIndex),
							"name":               info.SubSystemName,
							"node":               nodeName,
							"dcu_pod_namespace":  podShareInfo.Namespace,
							"dcu_pod_name":       podShareInfo.Pod,
							"container":          podShareInfo.Container,
							"vdcu_minor_number":  strconv.Itoa(info.VitualDeviceIndex),
							"vdcu_computer_unit": strconv.Itoa(info.ComputeUnitCount),
							"vdcu_memory_cap":    strconv.Itoa(info.MemoryCap),
						}
					} else if existsDyanamic {
						vdcuPrometheusLabels = prometheus.Labels{
							"device_id":          info.DeviceID,
							"minor_number":       strconv.Itoa(info.DeviceIndex),
							"name":               info.SubSystemName,
							"node":               nodeName,
							"dcu_pod_namespace":  podDynamicInfo.Namespace,
							"dcu_pod_name":       podDynamicInfo.Pod,
							"container":          podDynamicInfo.Container,
							"vdcu_minor_number":  strconv.Itoa(info.VitualDeviceIndex),
							"vdcu_computer_unit": strconv.Itoa(info.ComputeUnitCount),
							"vdcu_memory_cap":    strconv.Itoa(info.MemoryCap),
						}
					} else {
						vdcuPrometheusLabels = prometheus.Labels{
							"device_id":          info.DeviceID,
							"minor_number":       strconv.Itoa(info.DeviceIndex),
							"name":               info.SubSystemName,
							"node":               nodeName,
							"dcu_pod_namespace":  "",
							"dcu_pod_name":       "",
							"container":          "",
							"vdcu_minor_number":  strconv.Itoa(info.VitualDeviceIndex),
							"vdcu_computer_unit": strconv.Itoa(info.ComputeUnitCount),
							"vdcu_memory_cap":    strconv.Itoa(info.MemoryCap),
						}
					}

					vdcuTemp.With(vdcuPrometheusLabels).Set(info.Temperature)
					vdcuUsedMemoryBytes.With(vdcuPrometheusLabels).Set(float64(info.MemoryUsed))
					vdcuUtilizationRate.With(vdcuPrometheusLabels).Set(float64(info.UtilizationRate))
					vdcuSclk.With(vdcuPrometheusLabels).Set(info.Clk)

					// vdcu info
					glog.Infof("vdcu info : %v \n", info)
				}
			}

			time.Sleep(10 * time.Second)
		}
	}()
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func main() {
	flag.Parse()
	defer glog.Flush()
	_ = flag.Set("stderrthreshold", "INFO")

	glog.Infof("🚀 🚀 🚀  DCU exporter start ...")

	glog.Infof("Init ROCm smi: %v \n", dcgm.Init())
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
	registry.MustRegister(dcuComputeUnitCount)
	registry.MustRegister(dcuComputeUnitRemainingCount)
	registry.MustRegister(dcuMemoryRemaining)
	registry.MustRegister(dcuCE)
	registry.MustRegister(dcuUE)
	registry.MustRegister(vdcuSclk)
	registry.MustRegister(vdcuTemp)
	registry.MustRegister(vdcuUtilizationRate)
	registry.MustRegister(vdcuUsedMemoryBytes)

	recordMetrics()

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
