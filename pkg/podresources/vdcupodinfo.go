package podresources

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const configDir = "/etc/vdev"

type VitualDeviceInfo struct {
	//虚拟设备ID
	VitualDeviceIndex int

	//虚拟设备对应物理设备ID
	DeviceIndex int

	//虚拟设备对应物理设备号
	DeviceID string

	//设备名称
	SubSystemName string

	//温度
	Temperature float64

	//时钟频率
	Clk float64

	//虚拟设备计算单元分配量
	ComputeUnitCount int

	//虚拟设备内存分配量
	MemoryCap int

	// MemoryUsed 虚拟设备已使用的内存
	MemoryUsed int

	// UtilizationRate 虚拟设备的利用率
	UtilizationRate int
}

func GetVDCUPodInfo() (map[string]PodInfo, error) {
	result := make(map[string]PodInfo)

	err := filepath.Walk(configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.Contains(info.Name(), "_") {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading file %s: %w", path, err)
			}

			lines := strings.Split(string(content), "\n")
			fileStrings := strings.Split(info.Name(), "_")
			nameSpace := fileStrings[0]
			pod := fileStrings[1]

			for _, line := range lines {
				if line == "" {
					continue
				}

				parts := strings.Split(line, "_")
				if len(parts) != 3 {
					return fmt.Errorf("invalid line format in file %s: %s", path, line)
				}

				containerName := parts[0]
				vDCUID := parts[2]
				if err != nil {
					return fmt.Errorf("invalid vDCU ID in file %s: %s", path, line)
				}

				podInfo := PodInfo{
					Pod:       pod,
					Namespace: nameSpace,
					Container: containerName,
				}
				result[vDCUID] = podInfo
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", configDir, err)
	}

	return result, nil
}
