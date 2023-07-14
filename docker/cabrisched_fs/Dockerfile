# docker build -t cabrisched:0.1 -f docker/cabrisched_fs/Dockerfile .
FROM golang:1.20 as build
COPY gocode/go.mod /gocode/
COPY gocode/go.sum /gocode/
WORKDIR /gocode/
RUN go mod download
COPY gocode /gocode
WORKDIR /gocode/cabri
RUN CGO_ENABLED=0 go build -o cabri ./main.go

FROM alpine
RUN apk fix && \
    apk --no-cache --update add git && \
    adduser -D lambda && \
    touch /etc/cabri-pfile && \
    touch /etc/cabri-sfile
COPY --from=build /gocode/cabri/cabri /usr/local/bin
USER lambda
ENV PFILE /etc/cabri-pfile
ENV SFILE /etc/cabri-sfile
ENTRYPOINT /usr/local/bin/cabri schedule --pfile $PFILE --sfile $SFILE $*
