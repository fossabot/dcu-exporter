FROM ubuntu:22.04

WORKDIR /

COPY dcu-exporter-v2 /usr/local/bin/dcu-exporter-v2

COPY pkg/shim/lib/librocm_smi64.so /usr/lib/librocm_smi64.so.2.8

COPY pkg/shim/lib/libhydmi.so /usr/lib/libhydmi.so.1.4

RUN chmod +x /usr/local/bin/dcu-exporter-v2 \
    && ln -s /usr/lib/librocm_smi64.so.2.8 /usr/lib/librocm_smi64.so.2 \
    && ln -s /usr/lib/librocm_smi64.so.2 /usr/lib/librocm_smi64.so \
    && ln -s /usr/lib/libhydmi.so.1.4 /usr/lib/libhydmi.so.1 \
    && ln -s /usr/lib/libhydmi.so.1 /usr/lib/libhydmi.so

EXPOSE 16080

CMD  ["/usr/local/bin/dcu-exporter-v2", "-port=16080"]

# 在dockerfile中使用go build实在是太慢了，拉不下来依赖包，所以直接在外面运行编译了  ->   go build -o dcu-exporter-v2 main.go
# docker build -t dcu-exporter:v2.0.0.240718 .
# docker run --name dcu-exporter-v2 -d --privileged -v /etc/hostname:/etc/hostname -e LD_LIBRARY_PATH=$LD_LIBRARY_PATH  -p 16080:16080 dcu-exporter:v2.0.0.240718