// SPDX-License-Identifier: GPL-2.0-only OR MIT
//go:build ignore

#if defined(__x86_64__)
  #if KERNEL_VERSION > 511
    #include "vmlinux/x86_64/vmlinux_x86_64_6_11.h"
  #else
    #include "vmlinux/x86_64/vmlinux_x86_64_5_10.h"
  #endif
#else
#error "architecture not supported"
#endif
