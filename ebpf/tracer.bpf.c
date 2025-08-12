// SPDX-License-Identifier: GPL-2.0-only OR MIT
//go:build ignore

#include "vmlinux.h"
#include "maps.h"
#include "log_data.h"
#include "license.h"

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

/*
 *  Fill and send struct log_data to the ring buffer.
 */
static __always_inline void
kprobe_output(long unsigned int inode, dev_t dev, int mask)
{
  struct log_data data;
  
  __u64 pid_tgid = bpf_get_current_pid_tgid();
  __u64 uid_gid = bpf_get_current_uid_gid();

  data.tgid = pid_tgid >> 32;
  data.pid = (gid_t) pid_tgid;
  data.gid = uid_gid >> 32;
  data.uid = (gid_t) uid_gid;
  data.ino = inode;
  data.dev = dev;
  data.mask = mask;
  bpf_get_current_comm(data.comm, TASK_COMM_LEN);
		
  bpf_ringbuf_output(&rb, &data, sizeof(struct log_data), 0);
}

/*
 *  Probed function:
 *  int inode_permission(struct mnt_idmap *idmap,
 *	               	     struct inode *inode, int mask)
 *  Description: Check if accessing an inode is allowed
 */
SEC("kprobe/inode_permission")
int kprobe_inode_permission(struct pt_regs *ctx)
{
    struct mnt_idmap *idmap = (struct mnt_idmap*) PT_REGS_PARM1(ctx);
    struct inode *inode = (struct inode*) PT_REGS_PARM2(ctx);
    int mask = (int) PT_REGS_PARM3(ctx);

    long unsigned int ino = BPF_CORE_READ(inode, i_ino);
    dev_t dev = BPF_CORE_READ(inode, i_sb, s_dev);
    struct map_key key = {
      .inode = ino,
      .dev   = dev,
    };

    if (bpf_map_lookup_elem(&traced_inodes, &key))
    {
      kprobe_output(ino, dev, mask);
      return 0;
    }
    
    return 0;
}
