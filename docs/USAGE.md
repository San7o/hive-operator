# Usage

This document explains how to interact with the operator. You should
have the operator deployed first, to do that please read the
[DEVELOPMENT](./DEVELOPMENT.md) document for instructions.

Once you have the operator deployed, you can instruct It to log accesses
to files via HivePolicies. An **HivePolicy** is a custom kubernetes resource
that contains information about which file to trace and in which pods.
The operator will parse this policy every time one is added / removed / updated
and It will instruct the eBPF program to trace the right files.

An example HivePolicy is located in [config/samples/hive_v1alpha1_hivepolicy.yaml](../config/samples/hive_v1alpha1_hivepolicy.yaml).
More examples can be found in the same directory.

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
    label:
      security-level: high
```

This sample policy will trace the file `/secret.txt` in the pods with
name `nginx-pod` in the namespace `default` and with the label
`security-level=high`. If the file did not exist in the pod, It will
create It with `mode` permissions since `create` is set to true.

You can load it to the kubernetes cluster using the **apply** command
of [kubectl](https://kubernetes.io/docs/reference/kubectl/):

```bash
kubectl apply -f config/samples/hive_v1alpha1_hivepolicy.yaml
```

The operator will log some information when a policy is created /
deleted / updated.

Let's test this policy. First, we need to create a pod that matches
the `match` fields in the HivePolicy. This repository provides
an nginx pod in [config/samples/sample-nginx-pod.yaml](../config/samples/sample-nginx-pod.yaml)
with the right characteristics, you can load it with apply:

```bash
kubectl apply -f hack/k8s-manifests/sample-nginx-pod.yaml
```

After the pod has started, you can try to access the `/secret.txt` file
by executing a command inside it:

```bash
sudo kubectl exec -it nginx-pod -- cat /secret.txt
```

You should expect to see some logging information on the standard
output of one of the hive pods, like this:

```bash
 2025-04-28T09:32:48Z    INFO    New event    {"pid": 18116, "gid": 18116, "uid": 0, "gid": 0, "ino": 2736178, "mask": 36}
```

Note that only the leader pod in the kernel where the file resides
will output the information, so you may need to check the standard
output od all the hive pods.
