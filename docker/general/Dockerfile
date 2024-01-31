FROM alpine:3.19
RUN set -ex; \
	apkArch="$(apk --print-arch)"; \
	case "$apkArch" in \
		x86_64) arch='amd64' ;; \
		*) echo >&2 "error: unsupported architecture: $apkArch"; exit 1 ;; \
	esac; \
    version='v0.2.13'; \
	wget --quiet -O /tmp/cabri.tar.gz "https://github.com/t-beigbeder/otvl_cabri/releases/download/${version}/cabri-${version}-linux-${arch}.tar.gz"; \
	tar xzvf /tmp/cabri.tar.gz; \
    cp -p /tmp/cabri/cabri /usr/local/bin; \
	rm -fr /tmp/cabri.tar.gz /tmp/cabri; \
	chmod +x /usr/local/bin/cabri
COPY entrypoint.sh /
ENV SAUID 1001
ENV PFILE /etc/cabri-pfile
ENV VHOME /home/sa
ENV DATA /data
RUN touch $PFILE && \
    mkdir -p /home /docker-entrypoint-init.d && \
    chmod ugo+rwX /home && \
    chmod ugo+rx /entrypoint.sh

USER $SAUID
EXPOSE 3000
ENTRYPOINT ["/entrypoint.sh"]
CMD ["cabri", "webapi", "--haslog", "fsy+http://0.0.0.0:3000/${DATA}@data"]
