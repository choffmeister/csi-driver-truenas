# csi-driver-truenas

## Installation

```
export VERSION="v0.1.0-alpha.1"
kubectl apply -f https://github.com/choffmeister/csi-driver-truenas/releases/download/v$VERSION/manifest.yaml
kubectl apply -f - "
apiVersion: v1
kind: Secret
metadata:
  name: csi-driver-truenas-volumes
  namespace: kube-system
stringData:
  truenas-url: https://10.10.10.10
  truenas-api-key: "1-super-secret"
  truenas-tls-skip-verify: "true"
  truenas-parent-dataset: "tank/k8s"
  iscsi-base-iqn: "iqn.2005-10.org.freenas.ctl"
  iscsi-portal-ip: "10.10.10.10"
  iscsi-portal-port: "3260"
  iscsi-portal-id: "1"
  iscsi-initiator-id: "1"
"
```
