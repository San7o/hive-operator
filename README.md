# hive-operator

Hive is an eBPF-based file access monitoring Kubernetes operator.

# Basic Usage

You can specify a path to monitor and in which containers by
specifying an `HivePolicy`, for example:

```yaml
apiVersion: hive-operator.com/v1alpha1
kind: HivePolicy
metadata:
  labels:
    app.kubernetes.io/name: hive-operator
  name: hive-sample-policy
  namespace: hive-operator-system
spec:
  path: /secret.txt
  create: true
  mode: 444
  callback: "http://my-callback.com/alerts"
  match:
    pod: nginx-pod
    namespace: default
    matchLabels:
      security-level: high
```

The conditions under the `match` field will be matched via a logical
AND. All the match fields are optional; If none are specified, then
all containers are selected.

When a file gets accessed, the operator will generate an `HiveAlert`
and print the information to standard output in json format. The
following is an example alert:

```json
{
  "timestamp": "2025-07-25T08:14:22Z",
  "hive_policy_name": "hive-sample-policy",
  "metadata": {
    "path": "/secret.txt",
    "inode": 13667586,
    "mask": 34,
    "kernel_id": "fc9a30d5-6140-4dd1-b8ef-c638f19ebd71"
  },
  "pod": {
    "name": "nginx-pod",
    "namespace": "default",
    "contianer": {
      "id": "containerd://9d7df722223a4ad7f67f2afef5fbc0e263e23c7921011497f445e657fbced97e",
      "name": "nginx"
    }
  },
  "node": {
    "name": "hive-worker2"
  },
  "process": {
    "pid": 61116,
    "tgid": 61164
  }
}
```

If you specify a `callback` in the `HivePolicy`, then the data will be
sent to the URL of the callback through an HTTP POST request.

Please, read the [USAGE](./docs/USAGE.md) document to learn how to
use the operator in more detail.

# Development

The [DESIGN](./docs/DESIGN.md) document contains all the information
about the internals of the operator.

Please read the [DEVELOPMENT](./docs/DEVELOPMENT.md) document to build
and get started with Hive's
development. [EBPF-TESTING](./docs/EBPF-TESTING.md) has instructions
to build and test the eBPF program without running the kubernetes
operator. To run a local cluster, take a look at
[k8s-labs](./k8s-labs/README.md) or simply use the script
[registry-cluster.sh](./hack/registry-cluster.sh).

The [status](./docs/status.org) contains information about the current
status of development and future work.

## Current Limitations

Currently the only container runtime supported is containerd. The code
already uses an abstraction over container runtimes to easily
integrate more runtimes.
