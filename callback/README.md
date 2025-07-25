# Callback Service

This directory contains the callback container. This container hosts
the `/ingest` HTTP endpoint at default port `8090` and logs the body
of all the requests made to It. The `Hive` operator can use this
service to aggreate all logs from all nodes in a single palce.

On kubernetes, It registers a service at port `9376`, which can be
accessed by other pods via the url
`http://callback-service.hive-operator-system.svc.cluster.local:9376/ingest`.

The commands to build, containerize and deploy the application in a
kubernetes cluster are the same as the `ebpfdump` operator expect that
there is no code generation commands. Here is a quick summary:

```
make build      # Build
make docker     # Build doker image
make deploy     # Deploy on kubernetes
make undeploy   # Undeploy the application
```
