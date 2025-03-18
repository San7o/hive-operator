# Developement

To use an operator, you need a kubernetes cluster. To create one with
[kind](https://github.com/kubernetes-sigs/kind),
you can use the script `registry-cluster.sh` which will create
a cluster with one control node and one worker node. Additionally,
It sets up a local docker registry to push the operator's image
during developement.

```bash
make create-cluster-local
```

You can delete the cluster with `delete-cluster.sh` when you do not
want It anymore.

To generate the RBAC policies, run:
```bash
make generate
```

To create the CRD manifests, run:
```bash
make manifests
```

Note that you need to run the previous two commands only if you
actually changed something that needs to be regenerated.

To build everything inside a docker container, run:
```bash
make docker-build-local
make docker-push-local
```

Or in short:
```bash
make docker-local
```

This creates a docker image with the operator inside, and pushes It
int the local registry. Finally, you can deploy everything with:
```bash
make deploy
```

Check out [TESTING](./TESTING.md) for a guide on how to test the
application.

To delete the previously created cluster, run:
```bash
make delete-cluster-local
```

During developement, instead of building the container each time,
you can first try to compile with `make build` to check compiler
errors.

## Example Policy

You can apply the sample policy in `config/samples/hive_v1alpha1_hive.yaml`
```bash
sudo kubectl apply -f config/samples/hive_v1alpha1_hive.yaml
```
