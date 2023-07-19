# docker build -t cabrisched:0.1 -f docker/cabrisched_dev/Dockerfile .
# docker run --rm -p 3000:3000 -v /home/guest/secrets/cabri:/etc/cabri-pfile -v /home/guest/cabri_schedule.yaml:/etc/cabri-sfile.yaml cabrisched:0.1 -- --http
FROM golang:1.20 as build
COPY gocode/go.mod /gocode/
COPY gocode/go.sum /gocode/
WORKDIR /gocode/
RUN go mod download
COPY gocode /gocode
WORKDIR /gocode/cabri
RUN CGO_ENABLED=0 go build -o cabri ./main.go

FROM alpine
COPY --from=build /gocode/cabri/cabri /usr/local/bin
ENV PFILE /etc/cabri-pfile
ENV SFILE /etc/cabri-sfile.yaml
ENV SAUID 1001
ARG VHOME=/home/sa
RUN apk fix && \
    apk --no-cache --update add git
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
