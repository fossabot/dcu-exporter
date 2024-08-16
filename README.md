# DCU-Exporter

这个仓库包含了DCU-Exporter项目，它利用DCU的DCGM（Data Center GPGPU Manager）为Prometheus提供DCU指标导出功能。

## 文档

DCU-Exporter的官方文档可以在 [光合开发者社区](https://cancon.hpccube.com:65024/1/main) 上找到。

## 快速启动

### 物理机快速启动

前置条件：在安装DCU的节点上，启动dcu-exporter前需要安装dtk并使环境变量生效，或者将目标主机操作系统对应dtk的dtk/rocm_smi/lib目录下librocm_smi64.so.2.8动态链接库放置到目标主机/usr/lib64目录下，并在/usr/lib64目录创建指向librocm_smi64.so.2.8的软链接librocm_smi64.so.2和指向librocm_smi64.so.2的软链接librocm_smi64.so。

dcu-exporter启动直接运行可执行文件dcu-exporter-v2，dcu-exporter-v2支持启动参数和环境变量两种方式指定exporter服务端口。启动时添加-port参数指定端口，环境变量DCU_EXPORTER_LISTEN也可指定服务端口。优先启动参数指定，其次环境变量指定，最后默认16080。

```bash
./dcu-exporter-v2 -port=16080
```

使用curl命令来查看指标：

```bash
curl localhost:16080/metrics
```

你会看到如下的输出示例，显示了DCU内存规格、温度等信息：

```bash
# HELP dcu_memorycap_bytes dcu metrics of gauge
# TYPE dcu_memorycap_bytes gauge
dcu_memorycap_bytes{device_id="T8R1380013061601",minor_number="0",name="",node="dcunode3",pcieBus_number="0000:f6:00.0"} 3.4342961152e+10
dcu_memorycap_bytes{device_id="T8R1380019021101",minor_number="1",name="",node="dcunode3",pcieBus_number="0000:6a:00.0"} 3.4342961152e+10

# HELP dcu_temp dcu metrics of gauge
# TYPE dcu_temp gauge
dcu_temp{device_id="T8R1380013061601",minor_number="0",name="",node="dcunode3",pcieBus_number="0000:f6:00.0"} 46
dcu_temp{device_id="T8R1380019021101",minor_number="1",name="",node="dcunode3",pcieBus_number="0000:6a:00.0"} 47
...
```



### 容器快速启动

要在DCU节点上收集指标，只需启动dcu-exporter容器：

```bash
docker run --name dcu-exporter-v2 -d --privileged 
--device=/dev/kfd \
--device=/dev/mkfd \
--device=/dev/dri \
-v /opt/hyhal:/opt/hyhal \
-v /etc/hostname:/etc/hostname \
-e LD_LIBRARY_PATH=$LD_LIBRARY_PATH \
-p 16080:16080 dcu-exporter:v2.0.1
```

容器启动后，使用curl命令来查看指标：

```bash
curl localhost:16080/metrics
```

你会看到如下的输出示例，显示了DCU内存规格、温度等信息：

```bash
# HELP dcu_memorycap_bytes dcu metrics of gauge
# TYPE dcu_memorycap_bytes gauge
dcu_memorycap_bytes{device_id="T8R1380013061601",minor_number="0",name="",node="dcunode3",pcieBus_number="0000:f6:00.0"} 3.4342961152e+10
dcu_memorycap_bytes{device_id="T8R1380019021101",minor_number="1",name="",node="dcunode3",pcieBus_number="0000:6a:00.0"} 3.4342961152e+10

# HELP dcu_temp dcu metrics of gauge
# TYPE dcu_temp gauge
dcu_temp{device_id="T8R1380013061601",minor_number="0",name="",node="dcunode3",pcieBus_number="0000:f6:00.0"} 46
dcu_temp{device_id="T8R1380019021101",minor_number="1",name="",node="dcunode3",pcieBus_number="0000:6a:00.0"} 47
...
```

### Kubernetes快速启动

k8s集群一般以DaemonSet形式部署dcu-exporter，DaemonSet部署命令如下：

```bash
kubectl create -f dcu-exporter-v2.yaml
```

若需要自定义dcu-exporter 端口，可以修改 dcu-exporter-v2.yaml 文件中 contanierPort，Prometheus.io/port 和 service port。DaemonSet启动时以默认绑定Service，上述命令执行成功后，可以通过curl命令来查看指标：

```
curl -sL http://127.0.0.1:16080/metrics
```

你会看到如下的输出示例，显示了DCU内存规格、温度等信息；若有使用DCU的pod启动，指标输入将显示容器、POD等信息：

```bash
# HELP dcu_memorycap_bytes dcu metrics of gauge
# TYPE dcu_memorycap_bytes gauge
dcu_memorycap_bytes{device_id="T8R1380013061601",minor_number="0",name="",node="dcunode3",pcieBus_number="0000:f6:00.0"} 3.4342961152e+10
dcu_memorycap_bytes{device_id="T8R1380019021101",minor_number="1",name="",node="dcunode3",pcieBus_number="0000:6a:00.0"} 3.4342961152e+10

# HELP dcu_temp dcu metrics of gauge
# TYPE dcu_temp gauge
dcu_temp{device_id="T8R1380013061601",minor_number="0",name="",node="dcunode3",pcieBus_number="0000:f6:00.0"} 46
dcu_temp{device_id="T8R1380019021101",minor_number="1",name="",node="dcunode3",pcieBus_number="0000:6a:00.0"} 47
...
```

## 源码编译

为了构建 dcu-exporter，确保你具备如下前置条件：

1. 已安装 Golang 1.21 或更高版本
2. pkg模块存在正确版本的librocm_smi64.so动态链接库

编译命令如下：

```bash
git clone https://g.sugon.com/das/k8s-dcu.git
cd dcu-exporter-v2
export GOPROXY=https://goproxy.cn
export CGO_ENABLED=1
go mod tidy
go build -o dcu-exporter-v2 main.go
```

编译完成后，若需要通过docker启动，可通过如下命令进行dcu-exporter镜像制作：

```bash
docker build -t dcu-exporter:v2.0.0.240718 .
```

## 指标修改

//TODO

## Prometheus指标采集

Kubernetes方式部署dcu-exporter时，已经通过prometheus.io/scrape: 'true'，prometheus.io/port: &portStr '16080'，prometheus.io/path: 'metrics'注解开启Prometheus指标自动发现，只要k8s集群成功部署Prometheus即可自动从dcu-exporter采集指标。

## Grafana Dashboard

你可以在本项目grafana目录下获取dcu-exporter的Grafana Dashborad模板文件dcu-exporter-dashboard.json