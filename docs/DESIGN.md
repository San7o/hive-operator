# DESIGN

# Index

TODO

# Application Description

Hive is a kubernetes-native eBPF-based file access logging tool. The
user is able to select which file to trace; the application will
inform the user that an access to that file has happened.

**User story**: I, as the user of the application, want to log
all the `processes` that access the file `/etc/shadow` on containers
that are in the `security` namespace.

# Functional Requirements

This section describes the functional requirements of the application:

TODO

# Non functioncal requirements

TODO

# Design Considerations

The design of this application was made with the following considerations:
- the cluster runs on one or more linux operating systems
- one operating system may host one or more nodes
- each node's kubelet runs Its own container runtime
- pods may be scheduled and rescheduled in any node with any number
  of replicas

# Current limitations

Currently, the application supports `contaienrd` as the only supported
container runtime.

# Overview

This section contains a brief description of how the application
works, Its parts and how they interact with each other.

In particular, the application is composed of multiple components that
interact with eachother, kubernetes or the operating systems.
Those are:
- discover operator: collects identifying data about the files to check.
- loader operator: loads the eBPF program into the kernel.
- eBPF program: logs information when an actor interacts with any
    of the files selected by the user.
- databse: stores information about files to check.

A detailed description of the aforementioned components is given later.

## How dowe monitor accesses to files

Briefly, the end goal is to log when a file is accessed, that is,
when an actor interacts with It by opening, closing, writing, 
appending and so on. The application uses eBPF programs to monitor
accesses. More specifically, the eBPF program runs when
a certain kernel function is called, and It will check if said
function interacts with any of the files specified by the user. If
that is the case, It should log the information with additional
metadata such as which PID called the function. If you are new to
eBPF, you can think of them as simple programs that run inside
the kernel in a "trusted" way, hence they can access internal kernel
information that would only be available through kernel modules,
which are much more dangerous. The eBPF program needs to be loaded
and updated when the user changes configurations, therefore a loader 
and an updater are necessary, which are both done by the loader
operator, as well as logging information from the eBPF program to
stdout. Here I am oversimplifying and a more satisfying description
will be given later.

## How do we uniquely identify a file

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

## Kubernetes makes things harder

Now imagine all that we have said, but inside containers in a very
dynamic environment where things may change and break at any time.
There may be multiple operating systems so we need to load one eBPF
program for each one of them. We need to access inode numbers
of files inside containers that are only accessible inside a
kubernetes kubelet, pods (briefly, the applications in the cluster
if form of contaienrs) can be scheduled in any kubelet. All of this
needs to be handled carefully and greately increases the complexity
of the design.

## Example

An example deployment would look like the following:

![design-image](./images/overall-design.png)

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
up and now provide eBPF targets. A fundamental change to the eBPF
ecosystem was made with the introduction of the Bpf Type Format (BTF)
wich enables CO-RE (Compile Once, Run Everywhere). Using BTF will
enable the program to work on any kernel version. User space
provides eBPF programs to the kernel via the `bpf(2)` syscall, which will
verify that the program is correct (enforcing the previous limitations)
and will proceed to JIT compile it. People have been using eBPF for
tracing purposes. Moreover, eBPF programs can modify the kernel
innerworkings (such as the scheduler or cache policy) and, in recent
years, people are exploring its usage, even though the technology
is quite old (the original BPF was released in 1992).

## Overview of Kubernetes

TODO

# Detailed description

The different components are now described below:

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

Upon rescheduling of a pod, the process needs to be run again from
step 4. Upon rescheduling of the operator itself, the process
nedds to be run again from the start.

## Loader Operator

There is one loader operator for each running kernel. The loader
operator is responsible for the following operations:
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

## eBPF program

TODO

## Database

key = (pod, path)

TODO, how do we get database update events? Do we need to create a
custom event?
