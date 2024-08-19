# 编译
export GOPROXY=https://goproxy.cn
export CGO_ENABLED=1
go build -o dcu-exporter-v2 main.go
# 制作docker镜像
docker build -t dcu-exporter:v2.0.0.240718 .