FROM ubuntu:20.04

WORKDIR /

COPY dcu-exporter-v2 /usr/local/bin/dcu-exporter-v2

RUN chmod +x /usr/local/bin/dcu-exporter-v2

EXPOSE 16081

CMD  ["/usr/local/bin/dcu-exporter-v2"]

# 在dockerfile中使用go build实在是太慢了，拉不下来依赖包，所以直接在外面运行编译了  ->   go build -o dcu-exporter-v2 main.go
# docker build -t dcu-exporter:v2.0.1 .
# docker run --name dcu-exporter-v2 -d --privileged -v /opt/dtk-24.04:/opt/dtk-24.04 -v /etc/hostname:/etc/hostname -e LD_LIBRARY_PATH=$LD_LIBRARY_PATH  -p 16081:16081 dcu-exporter:v2.0.1
