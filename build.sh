# 编译
go build -o dcu-exporter-v2 main.go
# 制作docker镜像
docker build -t dcu-exporter:v2.0.1 .