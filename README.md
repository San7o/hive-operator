# hive-operator

Hive is an eBPF-based file access monitoring Kubernetes operator.

Currently, only pods using containerd runtime are supported.

# Usage

Specify a path to monitor by using a custom resource with the following
format:

```yaml
apiVersion: hive.com/v1alpha1
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
  match:
    pod: nginx-pod
    namespace: default
    labels:
      security-level: high
```

You can select match conditions to filter which pods to monitor for a
specific policy. All the match fields are optional. If none are
specified, all pods are selected. The operator will log accesses to
standard output with meaningful information, such as:

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
  "process": {
    "pid": 61116,
    "tgid": 61164
  }
}
```

Please, read the [USAGE](./docs/USAGE.md) document to learn how to
use the operator in more detail.

# Development

The [DESIGN](./docs/DESIGN.md) document contains all the information
about the internals of the operator.

Please read the [DEVELOPMENT](./docs/DEVELOPMENT.md) document to build
and get started with Hive's development.

The [status](./docs/status.org) contains information about the current
status of development and future work.
	
