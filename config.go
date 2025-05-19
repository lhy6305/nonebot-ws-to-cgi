package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var (
	clog_enable_debug           bool
	clog_enable_trace           bool
	clog_log_trace_to_file_only bool
	clog_logfile_path           string
	clog_enable_colored_output  bool

	ws_url        string
	ws_auth_token string

	ws_max_api_resp_queue_size int
	ws_max_pull_queue_size     int
	ws_max_push_queue_size     int

	cgi_program          string
	cgi_script_entry     string
	cgi_max_exec_time    time.Duration
	cgi_max_worker_count int

	http_listen_addr   string
	http_read_timeout  time.Duration
	http_write_timeout time.Duration

	http_api_timeout time.Duration
)

var (
	clog_logfile_handle *os.File = nil

	self_path = ""

	cgi_program_path = ""
	cgi_worker_sem   = make(chan struct{}, cgi_max_worker_count)
)

func config_parse_flags() {
	flag.BoolVar(&clog_enable_debug, "clog-enable-debug", true, "")
	flag.BoolVar(&clog_enable_trace, "clog-enable-trace", true, "")
	flag.BoolVar(&clog_log_trace_to_file_only, "clog-log-trace-to-file-only", false, "")
	flag.StringVar(&clog_logfile_path, "clog-logfile-path", "wss2cgi_log.txt", "")
	flag.BoolVar(&clog_enable_colored_output, "clog-enable-colored-output", true, "")

	flag.StringVar(&ws_url, "ws-url", "ws://example/path", "WebSocket服务器地址（支持wss）")
	flag.StringVar(&ws_auth_token, "ws-auth-token", "Bearer your_auth_token", "WebSocket认证令牌")

	flag.IntVar(&ws_max_api_resp_queue_size, "ws-max-api-resp-queue-size", 16, "")
	flag.IntVar(&ws_max_pull_queue_size, "ws-max-pull-queue-size", 16, "")
	flag.IntVar(&ws_max_push_queue_size, "ws-max-push-queue-size", 16, "")

	flag.StringVar(&cgi_program, "cgi-program", "php-cgi", "要运行的CGI解释器程序")
	flag.StringVar(&cgi_script_entry, "cgi-script-entry", "message_handler_entry.php", "CGI入口脚本相对路径")
	flag.DurationVar(&cgi_max_exec_time, "cgi-max-exec-time", 0*time.Second, "CGI进程最大执行时间")
	flag.IntVar(&cgi_max_worker_count, "cgi-max-worker-count", 16, "CGI最大工作进程数")

	flag.StringVar(&http_listen_addr, "http-listen-addr", "127.0.0.1:5700", "HTTP服务监听地址")
	flag.DurationVar(&http_read_timeout, "http-read-timeout", 0*time.Second, "HTTP读取超时时间")
	flag.DurationVar(&http_write_timeout, "http-write-timeout", 0*time.Second, "HTTP写入超时时间")

	flag.DurationVar(&http_api_timeout, "http-api-timeout", 30*time.Second, "HTTP API请求超时时间")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		fmt.Fprintf(flag.CommandLine.Output(), "wss2cgi program by ly65")
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		fmt.Fprintf(flag.CommandLine.Output(), "ly65-miao doesn't know how to use it, though...\r")
		fmt.Fprintf(flag.CommandLine.Output(), "                                               \n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags...]", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
	}

	flag.Parse()
}

func config_init() {

	config_parse_flags()

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

	custom_log("Info", "the cgi program path is %s", cgi_program_path)
	custom_log("Info", "the cgi program entry is %s", cgi_script_entry)
}
