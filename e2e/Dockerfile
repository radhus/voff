FROM alpine:latest

RUN apk add --no-cache \
        expect \
        mkinitfs \
        qemu-system-x86_64

WORKDIR /usr/src
COPY . .

CMD ["/usr/src/run.sh"]