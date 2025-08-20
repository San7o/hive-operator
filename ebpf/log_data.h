// SPDX-License-Identifier: GPL-2.0-only OR MIT
//go:build ignore

#ifndef _HIVE_DATA_H_
#define _HIVE_DATA_H_

#include "vmlinux.h"

#ifndef TASK_COMM_LEN
#define TASK_COMM_LEN 16
#endif

struct log_data {
	pid_t pid;                /* process id */
	gid_t tgid;               /* thread group id */
	uid_t uid;                /* user id */
	gid_t gid;                /* group id */
  dev_t dev;                /* device id */
	long unsigned int ino;    /* inode number */
	int mask;                 /* Octal representation of file permissions */
  char comm[TASK_COMM_LEN]; /* name of the executable of the task */
};

#endif // _HIVE_DATA_H_
