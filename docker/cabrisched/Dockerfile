# docker build -t cabrisched:0.2.2 -f docker/cabrisched/Dockerfile .
# docker run --rm -p 3000:3000 -v /home/guest/secrets/cabri:/etc/cabri-pfile -v /home/guest/cabri_schedule.yaml:/etc/cabri-sfile.yaml cabrisched:0.2.2 -- --http
FROM alpine
RUN apk fix && \
    apk --no-cache --update add git
RUN wget https://github.com/t-beigbeder/otvl_cabri/releases/download/v0.2.2/cabri-v0.2.2-linux-amd64.tar.gz
RUN [ "`sha256sum < cabri-v0.2.2-linux-amd64.tar.gz`" = "6bfd3055fe11e77b7dc679163125c37c1a9b51b04a4663b5a15a18844b27f6f8  -" ] && \
    tar xf cabri-v0.2.2-linux-amd64.tar.gz && \
    mv cabri/cabri /usr/local/bin && \
    chown root:root /usr/local/bin/cabri && \
    rm -r cabri-v0.2.2-linux-amd64.tar.gz cabri
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
