package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

func process_get_another_instance_pid() int32 {

	var self_pid int32 = int32(os.Getpid())

	self_proc, err := process.NewProcess(self_pid)
	if err != nil {
		custom_log("Warn", "Failed to get self process: %v", err)
		return 0
	}

	self_exe_path, err := process_get_normalized_path(self_proc)
	if err != nil {
		return 0
	}

	self_cmd_args, err := self_proc.CmdlineSlice()
	if err != nil {
		return 0
	}

	processes, err := process.Processes()
	if err != nil {
		return 0
	}

	for _, p := range processes {
		if p.Pid == self_pid {
			continue
		}

		exe_path, err := process_get_normalized_path(p)
		if err != nil || exe_path != self_exe_path {
			continue
		}

		args, err := p.CmdlineSlice()

		if err != nil || !process_is_cmd_arg_equal(self_cmd_args, args) {
			continue
		}

		return p.Pid
	}

	return 0
}

func process_get_normalized_path(p *process.Process) (string, error) {
	exe, err := p.Exe()
	if err != nil {
		return "", err
	}

	if absPath, err := filepath.Abs(exe); err == nil {
		exe = absPath
	}

	if runtime.GOOS == "windows" {
		exe = strings.ToLower(exe)
	}

	return filepath.Clean(exe), nil
}

func process_is_cmd_arg_equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
