package main

import (
	"fmt"
	"syscall"
)

func GetInode(Target string) (uint64, error) {

	fd, err := syscall.Open(Target, syscall.O_RDONLY, 444)
	if err != nil {
		return 0, fmt.Errorf("GetInode Error Open: %w", err)
	}
	defer syscall.Close(fd)

	var stat syscall.Stat_t
	err = syscall.Fstat(fd, &stat)
	if err != nil {
		return 0, fmt.Errorf("GetInode Error Fstat: %w", err)
	}

	return stat.Ino, nil
}
