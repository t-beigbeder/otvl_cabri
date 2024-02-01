FROM alpine:3.19
ARG ARCH="amd64"
ARG TAG="0.2.13"
RUN wget --quiet -O /tmp/cabri.tar.gz "https://github.com/t-beigbeder/otvl_cabri/releases/download/v${TAG}/cabri-v${TAG}-linux-${ARCH}.tar.gz" && \
	tar xzvf /tmp/cabri.tar.gz -C /usr/local/bin cabri/cabri --strip-components 1 && \
    rm -f /tmp/cabri.tar.gz && \
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
CMD cabri webapi --haslog fsy+http://0.0.0.0:3000${DATA}@data
