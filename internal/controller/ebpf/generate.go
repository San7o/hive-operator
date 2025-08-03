/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package ebpf

// To interact with the BPF program, we can (and should) generate
// some go bindings using the following command:
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -type log_data -tags linux -cflags "-D __${ARCH}__" bpf ../../../ebpf/tracer.bpf.c
