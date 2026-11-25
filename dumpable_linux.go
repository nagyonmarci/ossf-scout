package main

import "syscall"

// Scanners run as our uid, so without this they could read our secrets from
// /proc/<ppid>/environ or /proc/<ppid>/mem. A non-dumpable process has those
// files owned by root. Exec'd children become dumpable again, which is fine:
// their env is already scrubbed (scout.MinimalEnv).
func init() {
	// ponytail: best-effort hardening; prctl(PR_SET_DUMPABLE) does not fail on Linux in practice.
	_, _, _ = syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_SET_DUMPABLE, 0, 0)
}
