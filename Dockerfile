FROM ubuntu:22.04

WORKDIR /root

COPY dcu-exporter-v2 /usr/local/bin/dcu-exporter-v2

COPY pkg/shim/lib ./lib

RUN chmod +x /usr/local/bin/dcu-exporter-v2 \
    && ln -s /root/lib/librocm_smi64.so.2.8 /root/lib/librocm_smi64.so.2 \
    && ln -s /root/lib/librocm_smi64.so.2 /root/lib/librocm_smi64.so \
    && ln -s /root/lib/libhydmi.so.1.4 /root/lib/libhydmi.so.1 \
    && ln -s /root/lib/libhydmi.so.1 /root/lib/libhydmi.so

ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/root/lib

EXPOSE 16080

CMD  ["/usr/local/bin/dcu-exporter-v2", "-port=16080"]

# 在dockerfile中使用go build实在是太慢了，拉不下来依赖包，所以直接在外面运行编译了  ->   go build -o dcu-exporter-v2 main.go
# docker build -t dcu-exporter:v2.0.0.240718 .