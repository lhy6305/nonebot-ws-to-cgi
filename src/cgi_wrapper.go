package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/cgi"
	"strings"
	"time"
)

func call_cgi_wrapper(message []byte) {
	custom_log("Debug", "calling CGI with %d bytes", len(message))
	custom_log("Trace", "%v", string(message))

	cgi_worker_sem <- struct{}{}
	defer func() { <-cgi_worker_sem }()

	time_start := time.Now()
	resp := call_cgi(message)
	if resp == nil {
		custom_log("Error", "error in CGI call. see above for details")
		return
	}
	custom_log("Debug", "CGI processed in %s with status code %d", time.Since(time_start), resp.statusCode)
	if len(resp.body.Bytes()) > 0 {
		custom_log("Warn", "CGI output is not empty:\n%v", resp.String())
	} else {
		//custom_log("Trace", "CGI output:\n%v", resp.String())
	}
}

func call_cgi(data []byte) *responseRecorder {
	// Set up CGI environment variables
	env := []string{
		//"AUTH_TYPE=",
		//"CONTENT_LENGTH=" + strconv.Itoa(len(data)),
		//"CONTENT_TYPE=application/json",
		//"GATEWAY_INTERFACE=CGI/1.1",
		//"PATH_INFO=",
		//"PATH_TRANSLATED=",
		//"QUERY_STRING=",
		//"REMOTE_ADDR=127.0.0.1",
		//"REMOTE_HOST=",
		//"REMOTE_IDENT=",
		//"REMOTE_USER=",
		//"REQUEST_METHOD=POST",
		"SCRIPT_FILENAME=" + cgi_script_entry,
		//"SERVER_NAME=localhost",
		//"SERVER_PORT=80",
		//"SERVER_PROTOCOL=HTTP/1.1",
		//"SERVER_SOFTWARE=",
		"REDIRECT_STATUS=200", // Required for php-cgi program
	}

	handler := &cgi.Handler{
		Path: cgi_program_path,
		//Args: []string{filepath.Join(self_path, cgi_script_entry)},
		Dir: self_path,
		Env: env,
	}

	req, err := http.NewRequest("POST", "cgi://", bytes.NewReader(data))
	if err != nil {
		custom_log("Error", "failed to create request: %v", err)
		return nil
	}
	//req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(data))

	if cgi_max_exec_time > 0 {
		ctx, _ := context.WithTimeout(context.Background(), cgi_max_exec_time)
		req = req.WithContext(ctx)
	}

	resp := &responseRecorder{
		header: make(http.Header),
	}

	handler.ServeHTTP(resp, req)

	if resp.statusCode <= 0 {
		resp.statusCode = 200
	}

	return resp
}

type responseRecorder struct {
	header     http.Header
	body       bytes.Buffer
	statusCode int
}

func (r *responseRecorder) Header() http.Header {
	return r.header
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	return r.body.Write(data)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func (r *responseRecorder) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "\n----- begin of cgi output -----\n\n")
	fmt.Fprintf(&sb, "status code: %d\n", r.statusCode)
	for key, values := range r.header {
		fmt.Fprintf(&sb, "%s: %s\n", key, strings.Join(values, ", "))
	}
	fmt.Fprintf(&sb, "%q\n", r.body.String())
	fmt.Fprintf(&sb, "\n----- end of cgi output -----\n")
	return sb.String()
}
