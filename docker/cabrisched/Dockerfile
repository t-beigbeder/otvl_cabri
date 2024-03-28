# docker build -t cabrisched:0.2.2 -f docker/cabrisched/Dockerfile .
# docker run --rm -p 3000:3000 -v /home/guest/secrets/cabri:/etc/cabri-pfile -v /home/guest/cabri_schedule.yaml:/etc/cabri-sfile.yaml cabrisched:0.2.2 -- --http
FROM alpine:3.19
ARG ARCH="amd64"
ARG TAG="0.2.14"
RUN wget --quiet -O /tmp/cabri.tar.gz "https://github.com/t-beigbeder/otvl_cabri/releases/download/v${TAG}/cabri-v${TAG}-linux-${ARCH}.tar.gz" && \
	tar xzvf /tmp/cabri.tar.gz -C /usr/local/bin cabri/cabri --strip-components 1 && \
    rm -f /tmp/cabri.tar.gz && \
	chmod +x /usr/local/bin/cabri
ENV SAUID 1001
ENV PFILE /etc/cabri-pfile
ENV SFILE /etc/cabri-sfile.yaml
ARG VHOME=/home/sa
RUN touch $PFILE && \
    touch $SFILE && \
    mkdir -p /home && \
    chmod ugo+rwX /home && \
    echo "mkdir -p $VHOME" > /entrypoint.sh && \
    echo "export HOME=$VHOME" >> /entrypoint.sh && \
    echo "\$*" >> /entrypoint.sh && \
    chmod ugo+rx /entrypoint.sh
USER $SAUID
ENTRYPOINT /entrypoint.sh /usr/local/bin/cabri schedule --pfile $PFILE --sfile $SFILE $*
