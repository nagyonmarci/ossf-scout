package main

import (
	"syscall"
	"testing"
)

func TestProcessNotDumpable(t *testing.T) {
	if r, _, _ := syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_GET_DUMPABLE, 0, 0); r != 0 {
		t.Fatalf("PR_GET_DUMPABLE = %d, want 0", r)
	}
}
