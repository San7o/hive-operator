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
      matchAny:
      - pod: nginx-pod
        namespace: default
        matchLabels:
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

```
2025-08-02T16:51:19Z    INFO    Access Detected    {"HiveAlert": "{\"timestamp\":\"2025-08-02T16:51:19Z\",\"hive_policy_name\":\"hive-sample-policy\",\"metadata\":{\"path\":\"/secret.txt\",\"inode\":16256084,\"mask\":36,\"kernel_id\":\"2c147a95-23e5-4f99-a2de-67d5e9fdb502\"},\"pod\":{\"name\":\"nginx-pod\",\"namespace\":\"default\",\"container\":{\"id\":\"containerd://0c37512624823392d71e99a12011148db30ba7ea2a74fc7ff8bd5f85bc7b499c\",\"name\":\"nginx\"}},\"node\":{\"name\":\"hive-worker\"},\"process\":{\"pid\":176928,\"tgid\":176928,\"uid\":0,\"gid\":0,\"binary\":\"cat\",\"cwd\":\"\"}}"}
```

Note that only the leader pod in the kernel where the file resides
will output the information, so you may need to check the standard
output od all the hive pods.

You may have seen a message like this just above the alert:

```
Could not read /host/proc/176917/cwd while generating an HiveAlert, this can happen if the process terminated too quickly for the operator to react or the node is running in a container and procfs is not mounted in /host/real/proc
```

If you setup the cluster correctly, the problem here is that the `cat`
process died before the operator could read the information associated
with the process. Those missing informations will simply be empty
values in the output and do not cause the operator to break. Try to
keep the process alive and see that the warning does not appear and
you get additional information such as the current working directory
(cwd):

```
sudo kubectl exec -it nginx-pod -- cat /secret.txt -
```
