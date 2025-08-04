# Callback Service

This project contains a callback container for testing pupuses,
located in the
[callback/](https://github.com/San7o/hive-operator/tree/main/callback)
directory. This is used to collect all the `HiveAlert` in a central
place. This container hosts the `/ingest` HTTP endpoint at default
port `8090` and logs the body of all the requests made to it. The
`Hive` operator can use this service to aggreate all logs from all
nodes in a single place.

On kubernetes, it registers a service at port `9376`, which can be
accessed by other pods via the url
`http://callback-service.hive-operator-system.svc.cluster.local:9376/ingest`.

The commands to build, containerize and deploy the application in a
kubernetes cluster are the same as the `Hive` operator except that
there are no code generation commands. Here is a quick summary:

```bash
make build      # Build
make docker     # Build doker image
make deploy     # Deploy on kubernetes
make undeploy   # Undeploy the application
```
