# kube-wan-dns-refresh
Kubernetes CronJob to update DNS records for clusters behind a NAT with a dynamic WAN IP. This is useful for running "homelab" Kubernetes clusters sitting behind a home router.

Currently this project supports updating records in Route53. Credentials are expected to be provided via the environment. For all credential options, see https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials

## Usage

Using this project requires a JSON config file, structured as follows:

```json
{
  "route53records": {
   "A": [
    "my-k8s-hosted-website.com"
   ]
  }
}
```

To run the software one time:
```sh
./kube-wan-dns-refresh --config config.json
```

## Running as a CronJob

The real utility is to run this application at some frequency, to minimize latency between ISP IP changes and DNS updates. We can add the record config as a ConfigMap and mount it into a Pod run as a CronJob to do this. Example for running every 5 minutes:

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: kube-wan-dns-refresh
  namespace: default
spec:
  schedule: "*/5 * * * *" # Every 5 minutes
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: kube-wan-dns-refresh
            image: ghcr.io/wendtek/kube-wan-dns-refresh:latest
            imagePullPolicy: IfNotPresent
            args:
            - --config /config/records.json
            volumeMounts:
            - name: config-volume
              mountPath: /config
          restartPolicy: Never
          volumes:
            - name: config-volume
              configMap:
                name: dns-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: dns-config
  namespace: default
data:
  records.json: |
    {
      "route53records": {
       "A": [
        "my-k8s-hosted-website.com"
       ]
      }
    }
```
