//go:build none
// +build none

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var (
	clog_enable_debug           = true
	clog_enable_trace           = true
	clog_log_trace_to_file_only = false
	clog_logfile_path           = "wss2cgi_log.txt"
	clog_enable_colored_output  = true

	wss_url       = "wss://example/path"
	ws_auth_token = "Bearer your_auth_token"

	ws_max_api_resp_queue_size = 16
	ws_max_pull_queue_size     = 16
	ws_max_push_queue_size     = 16

	cgi_program          = "php-cgi"
	cgi_script_entry     = "message_handler_entry.php"
	cgi_max_exec_time    = 0 * time.Second
	cgi_max_worker_count = 16

	http_listen_addr   = "127.0.0.1:5700"
	http_read_timeout  = 0 * time.Second
	http_write_timeout = 0 * time.Second

	http_api_timeout = 10 * time.Second
)

var (
	clog_logfile_handle *os.File = nil

	self_path = ""

	cgi_program_path = ""
	cgi_worker_sem   = make(chan struct{}, cgi_max_worker_count)
)

func config_init() {
	var err error = nil

	self_path, err = os.Executable()
	if err != nil {
		custom_log("Fatal", "Failed to get self executable: %v", err)
		os.Exit(1)
	}
	self_path = filepath.Dir(self_path)

	if len(clog_logfile_path) > 0 {
		clog_logfile_handle, err = os.OpenFile(clog_logfile_path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			custom_log("Fatal", "Failed to open log file: %v", err)
			clog_logfile_handle = nil
			os.Exit(1)
		}
	}

	cgi_program_path, err = exec.LookPath(cgi_program)
	if err != nil {
		custom_log("Fatal", "Failed to lookup cgi executable: %v", err)
		os.Exit(1)
	}

	cgi_script_entry = filepath.Join(self_path, cgi_script_entry)
}
