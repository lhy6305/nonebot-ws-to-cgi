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
	clog_enable_debug           bool
	clog_enable_trace           bool
	clog_log_trace_to_file_only bool
	clog_logfile_path           string
	clog_enable_colored_output  bool

	wss_url       string
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

	flag.BoolVar(&clog_enable_debug, "clog-enable-debug", true, "启用调试日志")
	flag.BoolVar(&clog_enable_trace, "clog-enable-trace", true, "启用跟踪日志")
	flag.BoolVar(&clog_log_trace_to_file_only, "clog-log-trace-to-file-only", false, "仅将跟踪日志写入文件")
	flag.StringVar(&clog_logfile_path, "clog-logfile-path", "wss2cgi_log.txt", "日志文件路径")
	flag.BoolVar(&clog_enable_colored_output, "clog-enable-colored-output", true, "启用彩色日志输出")

	flag.StringVar(&wss_url, "wss-url", "wss://example/path", "WebSocket 服务器地址")
	flag.StringVar(&ws_auth_token, "ws-auth-token", "Bearer your_auth_token", "WebSocket 认证令牌")

	flag.IntVar(&ws_max_api_resp_queue_size, "ws-max-api-resp-queue-size", 16, "API响应队列的最大容量")
	flag.IntVar(&ws_max_pull_queue_size, "ws-max-pull-queue-size", 16, "拉取队列的最大容量")
	flag.IntVar(&ws_max_push_queue_size, "ws-max-push-queue-size", 16, "推送队列的最大容量")

	flag.StringVar(&cgi_program, "cgi-program", "php-cgi", "CGI 解释器路径（如 php-cgi）")
	flag.StringVar(&cgi_script_entry, "cgi-script-entry", "message_handler_entry.php", "CGI 入口脚本名称")
	flag.DurationVar(&cgi_max_exec_time, "cgi-max-exec-time", 0*time.Second, "CGI 进程最大执行时间（如 5s、1m）")
	flag.IntVar(&cgi_max_worker_count, "cgi-max-worker-count", 16, "CGI 最大工作进程数")

	flag.StringVar(&http_listen_addr, "http-listen-addr", "127.0.0.1:5700", "HTTP 服务监听地址")
	flag.DurationVar(&http_read_timeout, "http-read-timeout", 0*time.Second, "HTTP 读取超时时间（如 10s）")
	flag.DurationVar(&http_write_timeout, "http-write-timeout", 0*time.Second, "HTTP 写入超时时间（如 10s）")

	flag.DurationVar(&http_api_timeout, "http-api-timeout", 10*time.Second, "HTTP API 请求超时时间（如 5s）")
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
}
