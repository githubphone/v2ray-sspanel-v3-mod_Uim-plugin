FROM ubuntu:latest as builder

LABEL maintainer="Rico <rico93@outlook.com>"

ENV v2ray_version=4.18.0

RUN apt-get update
RUN apt-get install curl -y
RUN curl -L -o /tmp/go.sh https://raw.githubusercontent.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/master/install-release.sh
RUN chmod +x /tmp/go.sh
RUN /tmp/go.sh --version ${v2ray_version} --panelurl https://google.com --panelkey 55fUxDGFzH3n --nodeid 123456

FROM alpine:latest

COPY --from=builder /usr/bin/v2ray/v2ray /usr/bin/v2ray/
COPY --from=builder /usr/bin/v2ray/v2ctl /usr/bin/v2ray/
COPY --from=builder /usr/bin/v2ray/geoip.dat /usr/bin/v2ray/
COPY --from=builder /usr/bin/v2ray/geosite.dat /usr/bin/v2ray/
COPY config.json  /etc/v2ray/config.json

RUN set -ex && \
    apk --no-cache add ca-certificates && \
    update-ca-certificates && \
    apk add --update tzdata && \
    mkdir /var/log/v2ray/ && \
    chmod +x /usr/bin/v2ray/v2ctl && \
    chmod +x /usr/bin/v2ray/v2ray && \
    rm -rf /var/cache/apk/*

ENV TZ=Asia/Shanghai
ENV PATH /usr/bin/v2ray:$PATH
VOLUME  /var/log/v2ray/
WORKDIR /var/log/v2ray/

CMD     sed -i "s|\"port\": 2333,|\"port\": ${api_port},|"  "/etc/v2ray/config.json" &&\
        sed -i "s|\"https://google.com\"|\"${sspanel_url}\"|g" "/etc/v2ray/config.json" && \
        sed -i "s/\"55fUxDGFzH3n\"/\"${key}\"/g" "/etc/v2ray/config.json" && \
        sed -i "s/123456/${node_id}/g" "/etc/v2ray/config.json" && \
        sed -i "s/\"SpeedTestCheckRate\": 6/\"SpeedTestCheckRate\": ${speedtest}/g" "/etc/v2ray/config.json" && \
        v2ray -config=/etc/v2ray/config.json
