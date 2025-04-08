// SPDX-License-Identifier: GPL-2.0-only OR MIT
//go:build ignore

#ifndef _HIVE_DATA_H_
#define _HIVE_DATA_H_

#include "vmlinux.h"

struct log_data {
	pid_t pid;             /* process id */
	gid_t tgid;            /* thread group id */
	uid_t uid;             /* user id */
	gid_t gid;             /* group id */
	long unsigned int ino; /* inode number */
	int mask;              /* Octal representation of file permissions */
};

#endif // _HIVE_DATA_H_
