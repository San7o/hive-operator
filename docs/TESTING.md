# Testing

Currently, the operator will log inodes of the container's processes
of kubernete's Pods matching HivePolicies. 

To deploy the operator, run:
```bash
make deploy
```

To load a sample HivePolicy, you cal load the following:
```bash
kubectl apply -f config/samples/hive_v1alpha1_hive.yaml
```

To load a sample pod, you can load the following nginx image that
matches the filters of the sample policy:
```bash
kubectl apply -f config/samples/sample-nginx-pod.yaml
```

After deployment, wou should see the information
logged in the operator's standard output.

When testing, It may be useful to kill all the pods, which can be
done with:
```bash
make kill-pods-local
```
