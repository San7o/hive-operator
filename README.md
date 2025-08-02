# hive-operator

Hive is an eBPF-powered file access monitoring Kubernetes operator.

# Basic Usage

You can specify a path to monitor and in which containers by
creating an `HivePolicy`. The following is an example policy:

```yaml
apiVersion: hive-operator.com/v1alpha1
kind: HivePolicy
metadata:
  labels:
    app.kubernetes.io/name: hive-operator
  finalizers:
    - hive-operator.com/finalizer
  name: hive-sample-policy
  namespace: hive-operator-system
spec:
  traps:
  - path: /secret.txt
    create: true
    mode: 444
    callback: "http://my-callback.com/alerts"
    matchAny:
    - pod: nginx-pod
      namespace: default
      container-name: ".*"
      matchLabels:
        security-level: high
```

The match items under the `matchAny` field will be matched via a
logical OR, and each field in a match group is matched with a logical
AND. All the match fields are optional, but there must be at least
one match item unser `matchAny`.

When a file gets accessed, the operator will generate an `HiveAlert`
and print the information to standard output in json format. The
following is an example alert:

```json
{
  "timestamp": "2025-08-02T16:51:19Z",
  "hive_policy_name": "hive-sample-policy",
  "metadata": {
    "path": "/secret.txt",
    "inode": 16256084,
    "mask": 36,
    "kernel_id": "2c147a95-23e5-4f99-a2de-67d5e9fdb502"
  },
  "pod": {
    "name": "nginx-pod",
    "namespace": "default",
    "container": {
      "id": "containerd://0c37512624823392d71e99a12011148db30ba7ea2a74fc7ff8bd5f85bc7b499c",
      "name": "nginx"
    }
  },
  "node": {
    "name": "hive-worker"
  },
  "process": {
    "pid": 176928,
    "tgid": 176928,
    "uid": 0,
    "gid": 0,
    "binary": "cat",
    "cwd": "/"
  }
}
```

If you specify a `callback` in the `HivePolicy`, then the data will be
sent to the URL of the callback through an HTTP POST request.

Please, read the [USAGE](./docs/USAGE.md) document to learn how to use
the operator in more detail. You can find more examples in
[config/samples](./config/samples/).

## Quick deploy

To deploy the operator, simply run:

```bash
kubectl apply -f https://raw.githubusercontent.com/San7o/hive-operator/refs/heads/main/dist/install-remote.yaml
```

# Development

The [DESIGN](./docs/DESIGN.md) document contains all the information
about the internals of the operator.

Please read the [DEVELOPMENT](./docs/DEVELOPMENT.md) document to build
and get started with Hive's
development. [EBPF-TESTING](./docs/EBPF-TESTING.md) has instructions
to build and test the eBPF program without running the kubernetes
operator. To run a local cluster, take a look at
[k8s-lab](./k8s-lab/README.md) or simply use the script
[registry-cluster.sh](./hack/registry-cluster.sh).

The [status](./docs/status.org) contains information about the current
status of development and future work.

## Current Limitations

Currently the only container runtime supported is containerd. The code
already uses an abstraction over container runtimes to easily
integrate more runtimes.
