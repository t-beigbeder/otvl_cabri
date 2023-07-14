FROM alpine

RUN apk fix && \
    apk --no-cache --update add git && \
    wget https://github.com/t-beigbeder/otvl_cabri/releases/download/v0.2.1/cabri-v0.2.1-linux-amd64.tar.gz && \
    tar xf cabri-v0.2.1-linux-amd64.tar.gz && \
    mv cabri/cabri /usr/local/bin && \
    chown root:root /usr/local/bin/cabri && \
    rm -r cabri-v0.2.1-linux-amd64.tar.gz cabri && \
    adduser -D lambda && \
    touch /etc/cabri-pfile && \
    touch /etc/cabri-sfile

USER lambda
ENV PFILE /etc/cabri-pfile
ENV SFILE /etc/cabri-sfile
CMD /usr/local/bin/cabri schedule --pfile $PFILE --sfile $SFILE
