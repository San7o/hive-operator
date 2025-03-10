# Design Document

This document contains all the information necessary to understand
the application and Its implementation. It explains the design
decisions and how the different components interact with one another
and with other technologies. After reading this document, you will
gain a good understanding of how the application operates.

# Index

- [Overview](#overview)
  - [Application Description](#description)
  - [Components](#components)
  - [How to manitor accesses to files](#accesses)
  - [How to uniquely identify a file](#identify)
  - [Kubernetes makes things harder](#complications)
  - [Example](#example)
  - [Overview of eBPF](#ebpf-overview)
  - [Overview of Kubernetes](#kubernetes-overview)
- [Detailed description](#detailed-description)
  - [Design Considerations](#design-considerations)
  - [Discover Operator](#discover-operator)
  - [Loader Operator](#loader-operator)
  - [eBPF program](#ebpf-program)
  - [Database](#databse)
	- [Key-Value Store](#key-value)
	- [Relational](#relational)
	- [Custom Resource Definition](#crd)
- [Limitations](#limitations)

<a name="overview"></a>
# Overview

This section contains a brief description of how the application
works, Its parts and how they interact with each other.

<a name="description"></a>
## Application Description

Hive is a kubernetes-native eBPF-based file access logging tool. The
user is able to select which file to trace; the application will
inform the user that an access to that file has happened.

**User story**: I, as the user of the application, want to log
all the `processes` that access the file `/etc/shadow` on containers
that are in the `security` namespace.

<a name="components"></a>
## Components

The application is composed of multiple components that interact with
eachother, kubernetes or the operating systems. Those are:
- discover operator: collects identifying data about the files to check.
- loader operator: loads the eBPF program into the kernel.
- eBPF program: logs information when an actor interacts with any
    of the files selected by the user.
- databse: stores information about files to check.

A detailed description of the aforementioned components is given later
in this document.

<a name="accesses"></a>
## How to monitor accesses to files

Briefly, the end goal is to log when a file is accessed, that is,
when an actor interacts with It by opening, closing, writing, 
appending and so on. The application uses eBPF programs to monitor
accesses. More specifically, the eBPF program runs when
a certain kernel function is called, and It will check if said
function interacts with any of the files specified by the user. If
that is the case, It should log the information with additional
metadata such as which PID called the function.

If you are new to eBPF, you can think of them as simple programs that
run inside the kernel in a "trusted" way, hence they can access internal
kernel information that would only be available through kernel modules,
which are much more dangerous. The eBPF program needs to be loaded
and updated when the user changes configurations, therefore a loader 
and an updater are necessary, which are both done by the loader
operator, as well as logging information from the eBPF program to
stdout. Here I am oversimplifying and a more satisfying description
will be given later.

<a name="identify"></a>
## How to uniquely identify a file

To identify a certain file, we can use Its path name. There cannot
be two different files with the same path name but you can create
a symlink to a certain file. The path of the symlink will be different
from the path of the original file but the accessed data will be the
same. To solve this case, we are using the inode number instead.
Each file has an inode number and It is unique in the filesystem that
he lives in. This is how the kernel internally identifies files.
However, there is still an edge case where two filesystems may have
different files but with the same inode number between the two
filesystems. This happens because the inode number is an identifier
in a filesystem, but It has no meaning on another filesystem and
may aswell point to a different file. To solve this problem, we
can save both the inode number and the device id, which will be
different for each filesystem unless the filesystem has been bind
mounted. In this last case, from the user prespective, the binded
filesystem and the filesystem onto which the binded one is mounted
have different inode numbers so just this is enough.

In our application, the program that is responsible to get the inode
number and the device id is the *discover operator*. The loader
operator and the discover operator share information through a
databse. This is nececssary because:

<a name="complications"></a>
## Kubernetes makes things harder

Now, imagine all that we have just said, but inside containers in a very
dynamic environment where things may change and break at any time.
There may be multiple operating systems so we need to load one eBPF
program for each one of them. We need to access inode numbers
of files inside containers that are only accessible inside a
kubernetes kubelet, pods (briefly, the applications in the cluster
if form of contaienrs) can be scheduled in any kubelet. All of this
needs to be handled carefully and greately increases the complexity
of the design.

<a name="example"></a>
## Example

An example deployment would look like the following:

![design-image](./images/overall-design.png)

<a name="ebpf-overveiew"></a>
## Overview of eBPF

eBPF programs are programs that run inside the kernel in a controlled
environment. They can be hooked to traditional tracing systems such as
tracepoints, perf events and kprobes, and they will be executed when
the hook is triggered. An eBPF program has its own [instruction set](https://www.ietf.org/archive/id/draft-thaler-bpf-isa-00.html),
programs are limited to having at most 512 Bytes of stack size and 1
million instruction, loops are not allowed and functions can have up to 5
argumnets and only certain functions can be called.
Note that those (and other) limitations are changing rapidly and the
kernel verifier is getting always smarter and allows for softer limits.

Usually you do not write bytecode directly; instead you let a
compiler generate it for you. Traditionally, [BCC](https://github.com/iovisor/bcc)
is used to compile said programs, however, both LLVM and GCC have caught
up and now provide eBPF targets.

A fundamental change to the eBPF ecosystem was made with the
introduction of the Bpf Type Format (BTF)
wich enables CO-RE (Compile Once, Run Everywhere). Using BTF will
enable the program to work on any kernel version. User space
provides eBPF programs to the kernel via the `bpf(2)` syscall, which will
verify that the program is correct (enforcing the previous limitations)
and will proceed to JIT compile it.

People have been using eBPF for tracing purposes. Moreover, eBPF
programs can modify the kernel innerworkings (such as the scheduler or
cache policy) and, in recent years, people are exploring its usage more
broadly, even though the technology is quite old (the original BPF was
released in 1992).

<a name="kubernetes-overveiew"></a>
## Overview of Kubernetes

Kubernetes is a declarative container management software. The user
specifies the desired state of the containers and kubernetes will
try to update to the desired state. Applications should expect to be
interrupted at any time and failures should be handled gracefully.
Kubernetes can work with multiple contianer runtimes such as 
containerd or podman and interacts with the containers through their
runtime (for example, via a containerd client). Therefore, kubernetes
abstracts the management of single container, and focuses on the
scheduling and setting up of containers in (potentially) a cluster.

Each computing unit on the cluster is called a *node*. There are two
kinds of nodes: a worker node and the control pane. The former will
run the user's applications and services through contianers grouped
in *pods* inside a *kubelet* using a container runtime, the latter
forms the backbone of the kubernetes cluster and is responsbile for
central management of the workers. It is composed of the api server
(which the kubelet use to communicate with the control pane) etcd
(a highly-available key-value store), scheduler and a controller
manager which runs all of the above.

A common pattern found in kubernetes is the Operator, which is a
custom controller that manages some resources called *custom resources*
and extends the behaviour of the cluster.

<a name="detailed-description"></a>
# Detailed description

This section describes the application in more depth. It is
recommended to read the overview section first in order to get a
general understanding of the application before reading the details.

<a name="design-considerations"></a>
## Design Considerations

The design of this application was conducted considering the following:
- the cluster runs on one or more linux operating systems
- one operating system may host one or more nodes
- each node's kubelet runs Its own container runtime
- pods may be scheduled and rescheduled in any node with any number
  of replicas

The different components are now described below:

<a name="discover-operator"></a>
## Discover Operator

The *discover* operator is responsible for fetching files'
information such as inode number, and for storing them in the databse.
There must be one discover operator for each kubelet. This is
necessary because the operator has to talk directly to the container
runtime and access the pods' filesystem in order to read the inode.
Note that when referring to inodes we are technically referring to the
inode number.

In particular, the operator performs the following actions in sequence:
1. Identify the kernel instance: the operator will fetch an unique
       identifier for the running kernel (for example reading
	   `/proc/sys/kernel/random/boot_id`). This is needed because
	   the loader should tell the eBPF program only the inodes that
	   exist on the running kernel. In other words, and inode makes
	   sense only in the kernel where It runs. Therefore, the discover
	   operator needs to identify Its running kernel in order to
	   share the inodes with the right loader (there is one loader per
	   running kernel, more info below).
2. Read the operator's configuration to know which pods to filter
3. Initialize a connection with the container runtime of the
       kubelet where the operator lives. Talking to the container
	   engine is necessary to know which PID corresponds to which
	   container, and through the PID we can access the filesystem.
4. Read the (filtered) containers' PIDs
5. Read the inodes + device id: this is done by calling `stat(2)`
       on the file in `/proc/<PID>/root/<FILE_PATH>` with the path
	   provided in the operator's configuration and the PID from the
	   previous point. Both inode number and device id are needed to
	   uniquely identify a file because different filesystems may have
	   two different files but the same inode number, valid for Its
	   filesystem. To handle this cases, we need to save the device id.
	   If the file does not exist, the operato will handle this
	   gracefully and just log a message.
6. Add each inode + device id and other metadata to the database
       in the table identified by the running kernel identifier. The
	   databse will make sure that there is only one entry per
	   `(pod, path)` so, upon rescheduling, there will never be
	   duplicate data or outdated inodes.
7. Create a new event that the loader will catch, unless this is
		done automatically by the database implementation or the
		loader's implementation does not use a watch, that
		describes the change.

Upon rescheduling of a pod, the process needs to be run again from
step 4. Upon rescheduling of the operator itself, the process
nedds to be run again from the start.

<a name="loader-operator"></a>
## Loader Operator

The loader operator is responsible for the following operations:
- Log to stdout: the operator will read the output of the eBPF
    program from `/sys/kernel/debug/tracing/trace_pipe`, parses it,
	adds kubernetes information (such as the name of the pod
	corresponding to the inode and other information from the
	kubernetes topology) and logs everything to the standard output
	of the kubelet.
- Load the eBPF program.
- Update the eBPF map when there is a change in the database with
	added / deleted inode + device id.

Upon rescheduling of the operator, the eBPF program needs to be
reloaded (closed and loaded again).

There must be one loader operator for each running kernel. This is
necessary because the loader interacts directly with a running
kernel. It is usless to have multiple loaders in the same kernel,
but at least one is necessary to load the eBPF program. To implement
this, each node needs to fetch Its running kernel identifier
(see point 1 of the discover operator) and then run elections so that
only one node is elected per running kernel.

<a name="ebpf-program"></a>
## eBPF program

To check whether an actor has interacted with a file, the eBPF program
hooks to the function `inode_permission` through a keyprobe. This
function gets called everytime the permissions of an inode are checked,
which happends before any operation. It allows the eBPF program to log
when a permission is checked and with what rights, as well as who
tried to check the permissions. The eBPF program will log information
only if one of the inodes provided by the discover gets checked.
The loader will fetch those inodes from the databse and send them to
the eBPF program via a map, that is an array of pairs of `(inode number,
device id)`. The info is logged in `/sys/kernel/debug/tracing/trace_pipe`
and redirected to the kubelet's standard output by the loader's logger.

The eBPF program uses BTF types information to enable compile-once
run everywhere (CORE) which means that the ebpf program does not need
to be compiled each time It needs to be loaded, but can be compiled
only once and even shipped with the binaries of the application.

<a name="database"></a>
## Database

To communicacte data between the discover and the loader operator,
some sort of global data management is required. Each loader collects
data from Its kubelet and shares the data with the loader. We will
refer to this method as a databse, this description is necessary
because there are multiple valid methods to achive this and they will
all be referred as "database".
Theoretically, the inode data is only relevant to the operating
system that the discover operator lives in, hence the data can be
shared with only the loader operator that corresponds to that
operating system. Practically, this does not scale well because there
must be one database per operating system, therefore a global database
would be better.

Still, each loader needs to discriminate which inodes comes from which
operating system, so this information needs to be shared along with
the inode number and the device id. Due to the possibility of rapid
and frequent changes in the kubernetes cluster, the application must
ensure that updated information replaces the pre-existing one instead
of adding new entries. For example, suppose that a container named
`test` has the inode `1234` in the path `/home/root/secret` and the
databse has this information. If the container gets rescheduled, the
inode number may change to, say, `5678`. The databse should modify
the entry with the inode `1234` with `5678`. But how does the application
know which was the old inode number? It can identify it by It's path
and the container name, therefore those two are part of the key of
the databse.

Another reason for why we need a databse is to restore the eBPF
program's state when a loader dies and gets rescheduled. The loader
needs to access to all the information to send to the eBPF program.

There are multiple possible implementation of the databse, which are
discussed here:

<a name="key-value"></a>
### Key-Value Store

Data is saved as as key-value pairs and the data is retrieved by
using the key. Since the keys and values are composed of more than
one fields, those need to be serialized in some way, like comma
separated values or json. The key and the value fields would have the
following structure:

```
key = (pod-name, path-name)
value = (inode-number, device-id, running-kernel-identifier)
```

Kubernetes already uses `etcd` as a key value store, so there is no
need to run a different database. Querying the databse is costly, the
loader would iterate through all the tuples `(pod-name, path-name)`
and check if the key exists in the databse, then take only the inodes
that have the same running kernel identifier as the loader.

The loader may need to periodically query the database for updates.
A possibility to avoid this is to create a dummy custom resource
decifinition and update it when a value is written to the databse.
The loader would watch on that resource and query the database only
when the resource is updated.

<a name="relational"></a>
### Relational

A relational database would have the following schema:

```
|-----------|----------|--------------|-----------|-----------|
| path-name | pod-name | inode-number | device-id | kernel-id |
|-----------|----------|--------------|-----------|-----------|
```

There are many free solutions for distributed databases in kubernetes
such as [yugabyte-db](https://github.com/yugabyte/yugabyte-db).
To query the inodes, the loader would just have to select all the
rows that match Its kernel-id, at the cost of having a dedicated
database in the cluster, increasing resource utilization.

The relational database suffers from the same problem as the key
value store of needing to periodically query the databse, unless
the databse implementation provides some way to watch for changes
in the data.

<a name="crd"></a>
### Custom Resource Definition

A custom resource definition managed by the loader can be used to
store a row of the databse as a resource. The schema of the resource
would be the same as the databse and those would get updated
via regular CRUD operations on resources. The loader could watch
those resources so that It can react right after the change occurred.

In this case, the unique key constraint is not enforced by the system
so this must be handled manually by checking if a resource with the
same key exists before creating a new one.

<a name="limitations"></a>
# Limitations

- The usage of a databse increases resource utilization.
- The discover operator must implement different code for each
  supported container runtime.
