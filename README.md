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
    app.kubernetes.io/managed-by: kustomize
  name: hive-sample-policy
spec:
  monitors:
  - path: /secret.txt
    create: true
    mode: 444
    match:
      pod: my-pod
      namespace: hive-security
      label:
      - key: security-level
        value: high
```

You can select match conditions to filter which pods to monitor for a
specific policy. All the match fields are optional. If none are
specified, all pods are selected. The operator will log accesses to
standard output with meaningful information.

Please, read the [USAGE](./docs/USAGE.md) document to learn how to
use the operator in more detail.

# Development

Please read the [DEVELOPMENT](./docs/DEVELOPMENT.md) and
[TESTING](./docs/TESTING.md) documents to get started on Hive's
developement.

The [DESIGN](./docs/DESIGN.md) document contains all the information
about the internals of the operator.

The [status](./docs/status.org) contains information about the current
status of development and future work.
	
