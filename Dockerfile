FROM alpine:3.16
RUN apk add --no-cache blkid ca-certificates e2fsprogs e2fsprogs-extra
COPY --chmod=755 iscsiadm /sbin/iscsiadm
COPY csi-driver-truenas /bin/csi-driver-truenas
ENTRYPOINT ["/bin/csi-driver-truenas"]
