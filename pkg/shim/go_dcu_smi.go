package shim

/*
#cgo CFLAGS: -Wall -I/opt/dtk-24.04/rocm_smi/include/rocm_smi
#cgo LDFLAGS: -L/opt/dtk-24.04/rocm_smi/lib -lrocm_smi64 -Wl,--unresolved-symbols=ignore-in-object-files
#include <stdint.h>
#include <kfd_ioctl.h>
#include <rocm_smi64Config.h>
#include <rocm_smi.h>
*/
import "C"
import (
	"unsafe"
)

type RSMITemperatureMetric C.rsmi_temperature_metric_t

const (
	RSMI_TEMP_CURRENT        RSMITemperatureMetric = C.RSMI_TEMP_CURRENT
	RSMI_TEMP_FIRST          RSMITemperatureMetric = C.RSMI_TEMP_FIRST
	RSMI_TEMP_MAX            RSMITemperatureMetric = C.RSMI_TEMP_MAX
	RSMI_TEMP_MIN            RSMITemperatureMetric = C.RSMI_TEMP_MIN
	RSMI_TEMP_MAX_HYST       RSMITemperatureMetric = C.RSMI_TEMP_MAX_HYST
	RSMI_TEMP_MIN_HYST       RSMITemperatureMetric = C.RSMI_TEMP_MIN_HYST
	RSMI_TEMP_CRITICAL       RSMITemperatureMetric = C.RSMI_TEMP_CRITICAL
	RSMI_TEMP_CRITICAL_HYST  RSMITemperatureMetric = C.RSMI_TEMP_CRITICAL_HYST
	RSMI_TEMP_EMERGENCY      RSMITemperatureMetric = C.RSMI_TEMP_EMERGENCY
	RSMI_TEMP_EMERGENCY_HYST RSMITemperatureMetric = C.RSMI_TEMP_EMERGENCY_HYST
	RSMI_TEMP_CRIT_MIN       RSMITemperatureMetric = C.RSMI_TEMP_CRIT_MIN
	RSMI_TEMP_CRIT_MIN_HYST  RSMITemperatureMetric = C.RSMI_TEMP_CRIT_MIN_HYST
	RSMI_TEMP_OFFSET         RSMITemperatureMetric = C.RSMI_TEMP_OFFSET
	RSMI_TEMP_LOWEST         RSMITemperatureMetric = C.RSMI_TEMP_LOWEST
	RSMI_TEMP_HIGHEST        RSMITemperatureMetric = C.RSMI_TEMP_HIGHEST
	RSMI_TEMP_LAST           RSMITemperatureMetric = C.RSMI_TEMP_LAST
)

type RSMIMemoryType C.rsmi_memory_type_t

const (
	RSMI_MEM_TYPE_FIRST    RSMIMemoryType = C.RSMI_MEM_TYPE_FIRST
	RSMI_MEM_TYPE_VRAM     RSMIMemoryType = C.RSMI_MEM_TYPE_VRAM
	RSMI_MEM_TYPE_VIS_VRAM RSMIMemoryType = C.RSMI_MEM_TYPE_VIS_VRAM
	RSMI_MEM_TYPE_GTT      RSMIMemoryType = C.RSMI_MEM_TYPE_GTT
	RSMI_MEM_TYPE_LAST     RSMIMemoryType = C.RSMI_MEM_TYPE_LAST
)

type RSMIFrequencies struct {
	NumSupported uint32
	Current      uint32
	Frequency    [32]uint64
}
type RSMIPcieBandwidth struct {
	TransferRate RSMIFrequencies
	Lanes        [32]uint32
}

type RSMIClkType C.rsmi_clk_type_t

const (
	RSMI_CLK_TYPE_SYS   RSMIClkType = C.RSMI_CLK_TYPE_SYS
	RSMI_CLK_TYPE_FIRST RSMIClkType = C.RSMI_CLK_TYPE_FIRST
	RSMI_CLK_TYPE_DF    RSMIClkType = C.RSMI_CLK_TYPE_DF
	RSMI_CLK_TYPE_DCEF  RSMIClkType = C.RSMI_CLK_TYPE_DCEF
	RSMI_CLK_TYPE_SOC   RSMIClkType = C.RSMI_CLK_TYPE_SOC
	RSMI_CLK_TYPE_MEM   RSMIClkType = C.RSMI_CLK_TYPE_MEM
	RSMI_CLK_TYPE_LAST  RSMIClkType = C.RSMI_CLK_TYPE_LAST
	RSMI_CLK_INVALID    RSMIClkType = C.RSMI_CLK_INVALID
)

/****************************************** Initialize *********************************************/

// GO_rsmi_init 初始化rocm_smi
func GO_rsmi_init() uint {
	return uint(C.rsmi_init(0))
}

// GO_rsmi_shutdown 关闭rocm_smi
func GO_rsmi_shutdown() uint {
	return uint(C.rsmi_shut_down())
}

/****************************************** Identifier *********************************************/

// GO_rsmi_num_monitor_devices 获取gpu数量 *
func GO_rsmi_num_monitor_devices() int {
	var p C.uint
	C.rsmi_num_monitor_devices(&p)
	return int(p)
}

// TODO GO_rsmi_dev_sku_get 获取设备sku
func GO_rsmi_dev_sku_get(dvInd int) string {
	var sku C.ushort
	C.rsmi_dev_sku_get(C.uint(dvInd), &sku)
	return string(sku)
}

// GO_rsmi_dev_vendor_id_get 获取设备供应商id
func GO_rsmi_dev_vendor_id_get(i C.uint) uint {
	var vid C.ushort
	C.rsmi_dev_vendor_id_get(i, &vid)
	return uint(vid)
}

// GO_rsmi_dev_id_get 获取设备id
func GO_rsmi_dev_id_get(i int) uint {
	var iid C.ushort
	C.rsmi_dev_id_get(C.uint(i), &iid)
	return uint(iid)
}

// GO_rsmi_dev_name_get 获取设备名称
func GO_rsmi_dev_name_get(i C.uint) string {
	name := make([]C.char, uint32(256))
	C.rsmi_dev_name_get(i, (*C.char)(unsafe.Pointer(&name[0])), 256)
	nameStr := C.GoString((*C.char)(unsafe.Pointer(&name[0])))
	return nameStr
}

// GO_rsmi_dev_brand_get 获取设备品牌名称
func GO_rsmi_dev_brand_get(i C.uint) string {
	brand := make([]C.char, uint32(256))
	C.rsmi_dev_brand_get(i, (*C.char)(unsafe.Pointer(&brand[0])), 256)
	result := C.GoString((*C.char)(unsafe.Pointer(&brand[0])))
	return result
}

// GO_rsmi_dev_vendor_name_get 获取设备供应商名称
func GO_rsmi_dev_vendor_name_get(i C.uint) string {
	bname := make([]C.char, uint32(256))
	C.rsmi_dev_vendor_name_get(i, (*C.char)(unsafe.Pointer(&bname[0])), 80)
	result := C.GoString((*C.char)(unsafe.Pointer(&bname[0])))
	return result
}

// GO_rsmi_dev_vram_vendor_get 获取设备显存供应商名称
func GO_rsmi_dev_vram_vendor_get(i C.uint) string {
	bname := make([]C.char, uint32(256))
	C.rsmi_dev_vram_vendor_get(i, (*C.char)(unsafe.Pointer(&bname[0])), 80)
	result := C.GoString((*C.char)(unsafe.Pointer(&bname[0])))
	return result
}

// GO_rsmi_dev_serial_number_get 获取设备序列号 *
func GO_rsmi_dev_serial_number_get(i int) string {
	serialNumber := make([]C.char, uint32(256))
	C.rsmi_dev_serial_number_get(C.uint(i), (*C.char)(unsafe.Pointer(&serialNumber[0])), 256)
	result := C.GoString((*C.char)(unsafe.Pointer(&serialNumber[0])))
	return result
}

// GO_rsmi_dev_subsystem_id_get 获取设备子系统id
func GO_rsmi_dev_subsystem_id_get(i C.uint) int {
	var id C.ushort
	C.rsmi_dev_subsystem_id_get(i, &id)
	return int(id)
}

// GO_rsmi_dev_subsystem_name_get 获取设备子系统名称 *
func GO_rsmi_dev_subsystem_name_get(i int) string {
	subSystemName := make([]C.char, uint32(256))
	C.rsmi_dev_subsystem_name_get(C.uint(i), (*C.char)(unsafe.Pointer(&subSystemName[0])), 256)
	result := C.GoString((*C.char)(unsafe.Pointer(&subSystemName[0])))
	return result
}

// GO_rsmi_dev_drm_render_minor_get 获取设备drm次编号
func GO_rsmi_dev_drm_render_minor_get(i C.uint) int {
	var id C.uint
	C.rsmi_dev_drm_render_minor_get(i, &id)
	return int(id)
}

// GO_rsmi_dev_unique_id_get 获取设备唯一id
func GO_rsmi_dev_unique_id_get(dvInd int) int64 {
	var uniqueId C.ulong
	C.rsmi_dev_unique_id_get(C.uint(dvInd), &uniqueId)
	return int64(uniqueId)
}

/****************************************** PCIe *********************************************/

// GO_rsmi_dev_pci_id_get 获取唯一pci设备标识符
func GO_rsmi_dev_pci_id_get(dvInd int) int64 {
	var bdfid C.ulong
	C.rsmi_dev_pci_id_get(C.uint(dvInd), &bdfid)
	return int64(bdfid)
}

// GO_rsmi_dev_pci_bandwidth_get 获取可用的pcie带宽列表
func GO_rsmi_dev_pci_bandwidth_get(dvInd C.uint) RSMIPcieBandwidth {
	var bandwidth C.rsmi_pcie_bandwidth_t
	C.rsmi_dev_pci_bandwidth_get(C.uint32_t(dvInd), &bandwidth)
	rsmiPcieBandwidth := RSMIPcieBandwidth{
		TransferRate: RSMIFrequencies{
			NumSupported: uint32(bandwidth.transfer_rate.num_supported),
			Current:      uint32(bandwidth.transfer_rate.current),
			Frequency:    *(*[32]uint64)(unsafe.Pointer(&bandwidth.transfer_rate.frequency)),
		},
		Lanes: *(*[32]uint32)(unsafe.Pointer(&bandwidth.lanes)),
	}
	return rsmiPcieBandwidth
}

// GO_rsmi_dev_pci_throughput_get 获取pcie流量信息
func GO_rsmi_dev_pci_throughput_get(dvInd int) (sent, received, maxPktSz int64) {
	var csent, creceived, cmaxpktsz C.ulong
	C.rsmi_dev_pci_throughput_get(C.uint(dvInd), &csent, &creceived, &cmaxpktsz)
	sent, received, maxPktSz = int64(csent), int64(creceived), int64(cmaxpktsz)
	return sent, received, maxPktSz
}

/****************************************** Power *********************************************/

// GO_rsmi_dev_power_ave_get 获取设备平均功耗
func GO_rsmi_dev_power_ave_get(dvInd int, senserId int) int64 {
	var power C.ulong
	C.rsmi_dev_power_ave_get(C.uint(dvInd), C.uint(senserId), &power)
	return int64(power)
}

// GO_rsmi_dev_power_cap_get 获取设备功率上限
func GO_rsmi_dev_power_cap_get(dvInd int, senserId int) int64 {
	var power C.ulong
	C.rsmi_dev_power_cap_get(C.uint(dvInd), C.uint(senserId), &power)
	return int64(power)
}

// GO_rsmi_dev_power_cap_range_get 获取设备功率有效值范围
func GO_rsmi_dev_power_cap_range_get(dvInd int, senserId int) (max, min int64) {
	var cmax, cmin C.ulong
	C.rsmi_dev_power_cap_range_get(C.uint(dvInd), C.uint(senserId), &cmax, &cmin)
	max, min = int64(cmax), int64(cmin)
	return max, min
}

/****************************************** Memory *********************************************/

// GO_rsmi_dev_memory_total_get 获取设备内存总量 *
func GO_rsmi_dev_memory_total_get(dvInd int, memoryType RSMIMemoryType) int64 {
	var total C.ulong
	C.rsmi_dev_memory_total_get(C.uint(dvInd), C.rsmi_memory_type_t(memoryType), &total)
	return int64(total)
}

// GO_rsmi_dev_memory_usage_get 获取当前设备内存使用情况 *
func GO_rsmi_dev_memory_usage_get(dvInt int, memoryType RSMIMemoryType) int64 {
	var used C.ulong
	C.rsmi_dev_memory_usage_get(C.uint(dvInt), C.rsmi_memory_type_t(memoryType), &used)
	return int64(used)
}

// GO_rsmi_dev_memory_busy_percent_get 获取设备内存使用的百分比
func GO_rsmi_dev_memory_busy_percent_get(dvInt int) int {
	var busyPercent C.uint
	C.rsmi_dev_memory_busy_percent_get(C.uint(dvInt), &busyPercent)
	return int(busyPercent)
}

/****************************************** Physical State *********************************************/

// GO_rsmi_dev_temp_metric_get 获取设备的温度度量值 *
func GO_rsmi_dev_temp_metric_get(dvInd int, sensor_type C.uint, metric RSMITemperatureMetric) int64 {
	var temperature C.long
	C.rsmi_dev_temp_metric_get(C.uint(dvInd), sensor_type, C.rsmi_temperature_metric_t(metric), &temperature)
	return int64(temperature)
}

/****************************************** Performance *********************************************/

// GO_rsmi_dev_busy_percent_get 获取设备设备忙碌时间百分比 *
func GO_rsmi_dev_busy_percent_get(dvInd int) int {
	var busyPercent C.uint
	C.rsmi_dev_busy_percent_get(C.uint(dvInd), &busyPercent)
	return int(busyPercent)
}

// GO_rsmi_dev_gpu_clk_freq_get 获取设备系统时钟速度列表
func GO_rsmi_dev_gpu_clk_freq_get(dvInd int, clkType RSMIClkType) RSMIFrequencies {
	var rsmiFrequencies C.rsmi_frequencies_t
	C.rsmi_dev_gpu_clk_freq_get(C.uint(dvInd), C.rsmi_clk_type_t(clkType), &rsmiFrequencies)
	rf := RSMIFrequencies{
		NumSupported: uint32(rsmiFrequencies.num_supported),
		Current:      uint32(rsmiFrequencies.current),
		Frequency:    *(*[32]uint64)(unsafe.Pointer(&rsmiFrequencies.frequency)),
	}
	return rf
}

/****************************************** Version *********************************************/

/****************************************** Error *********************************************/

/****************************************** System *********************************************/

/****************************************** XGMI *********************************************/

/****************************************** TOPO *********************************************/

/****************************************** Supported *********************************************/

/****************************************** Event *********************************************/
