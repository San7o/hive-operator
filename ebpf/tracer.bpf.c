// SPDX-License-Identifier: GPL-2.0-only OR MIT
//go:build ignore

#include "vmlinux.h"
#include "maps.h"
#include "log_data.h"
#include "license.h"

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

void check(const long unsigned int* ino, const int* mask, const __u32 i) {
	long unsigned int *valp;
	valp = bpf_map_lookup_elem(&traced_inodes, &i);
	if (valp && *valp == *ino) {
		__u64 pid_tgid = bpf_get_current_pid_tgid();
		__u64 uid_gid = bpf_get_current_uid_gid();

		struct log_data data;
		data.pid = pid_tgid >> 32;
		data.tgid = (gid_t) pid_tgid;
		data.uid = uid_gid >> 32;
		data.gid = (gid_t) uid_gid;
		data.ino = *ino;
		data.mask = *mask;
		
		bpf_ringbuf_output(&rb, &data, sizeof(struct log_data), 0);
	}
	return;
}

SEC("kprobe/inode_permission")
int kprobe_inode_permission(struct pt_regs *ctx)
// Probed function:
// int inode_permission(struct mnt_idmap *idmap,
//		     struct inode *inode, int mask)
// Description: Check if accessing an inode is allowed
{
    struct mnt_idmap *idmap = (struct mnt_idmap*) BPF_CORE_READ(ctx, di);
    struct inode *inode = (struct inode*) BPF_CORE_READ(ctx, si);
    int mask = (int) BPF_CORE_READ(ctx, dx);
    long unsigned int ino = BPF_CORE_READ(inode, i_ino);

    #pragma unroll
		for (__u32 i = 0; i < MAP_MAX_ENTRIES; ++i) {
			check(&ino, &mask, i);
		}

    // Test data
    struct log_data data;
		data.pid = 123;
		data.tgid = 123;
		data.uid = 123;
		data.gid = 123;
		data.ino = ino;
		data.mask = mask;
    bpf_ringbuf_output(&rb, &data, sizeof(struct log_data), 0);
    
    return 0;
}
