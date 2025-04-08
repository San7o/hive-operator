// SPDX-License-Identifier: GPL-2.0-only OR MIT
//go:build ignore

#ifndef _HIVE_MAPS_H_
#define _HIVE_MAPS_H_

#include "vmlinux.h"
#include "log_data.h"
#include <bpf/bpf_helpers.h>

#define MAP_MAX_ENTRIES 1024

struct {
  __uint(type, BPF_MAP_TYPE_ARRAY);
  __type(key, __u32);
  __type(value, long unsigned int);
  __uint(max_entries, MAP_MAX_ENTRIES);
} traced_inodes SEC(".maps"); 

struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 1 << 24);
  __type(value, struct log_data);
} rb SEC(".maps");

#endif // _HIVE_MAPS_H_
