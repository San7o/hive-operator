# hive-operator

Hive is an eBPF-based file access monitoring Kubernetes operator.

Currently, only pods using containerd runtime are supported.

# Usage

Specify a path to monitor by using a custom resource with the following
format:
```yaml
apiVersion: hive.dynatrace.com/v1alpha1
kind: HivePolicy
metadata:
  name: hive-sample-policy
spec:
  monitors:
  - path: /etc/passwd
    create: true
    match:
	  pod: my-pod
	  namespace: hive-security
	  label: security
```

You can select match conditions to filter which pods to monitor
for a specific policy. If none are specified, all pods are considered.
The operator will log accesses to standard output with meaningful
information.

# Development

Please read the [DEVELOPMENT](./docs/DEVELOPMENT.md) and
[TESTING](./docs/TESTING.md) documents to get started on Hive's
developement. Read the [DESIGN](./docs/DESIGN.md) document to learn
how the operator works.

# Status

Pleas read the [status](./docs/status.org) document for information
about future work.
	
