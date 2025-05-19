package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	http_api_pending_pool       = make(map[string](chan []byte), ws_max_api_resp_queue_size)
	http_api_pending_pool_mutex = sync.Mutex{}
)

// async call
func http_server_loop() {
	server := &http.Server{
		Addr:         http_listen_addr,
		Handler:      http.HandlerFunc(http_handler),
		ReadTimeout:  http_read_timeout,
		WriteTimeout: http_write_timeout,
	}

	for {
		custom_log("Info", "Starting HTTP server on %s", http_listen_addr)
		err := server.ListenAndServe()
		custom_log("Error", "HTTP server failed: %v", err)
		time.Sleep(500 * time.Millisecond)
	}
}

func http_handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Vary", "Origin")

	// allow preflight requests
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(200)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")

	if r.Method != http.MethodPost {
		http_write_error(w, 405, "request method not allowed")
		return
	}

	if len(ws_auth_token) > 0 {
		authToken := r.Header.Get("Authorization")

		if authToken != ws_auth_token {
			http_write_error(w, 403, "invalid auth headers")
			custom_log("Warn", "Invalid auth token: %s", authToken)
			return
		}
	}

	action := strings.Trim(r.URL.Path, "/ \t\r\n")
	if len(action) <= 0 {
		http_write_error(w, 400, "bad request: you should specify action in request path")
		return
	}

	// if the client send Expect header, the http package will handle it automatically
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http_write_error(w, 400, "bad request: failed to read request body")
		custom_log("Error", "Failed to read body: %v", err)
		return
	}

	if ws_conn == nil {
		http_write_error(w, 503, "service unavailable: backend websocket not connected")
		custom_log("Warn", "http server: ws not connected")
		return
	}

	var params map[string]interface{}
	if err := json.Unmarshal(body, &params); err != nil {
		http_write_error(w, 400, "bad request: failed to parse request body as json: "+err.Error())
		custom_log("Error", "http server: invalid JSON input: %v", err)
		return
	}
	unique_id := gen_unique_id()

	packed_json_msg := make(map[string]interface{})
	packed_json_msg["action"] = action
	packed_json_msg["params"] = params
	packed_json_msg["echo"] = unique_id

	body, err = json.Marshal(packed_json_msg)
	if err != nil {
		http_write_error(w, 500, "internal server error: failed to marshal json: "+err.Error())
		custom_log("Error", "http server: Failed to marshal JSON: %v", err)
		return
	}

	resp_chan := make(chan []byte, 1)

	http_api_pending_pool_mutex.Lock()
	http_api_pending_pool[unique_id] = resp_chan
	http_api_pending_pool_mutex.Unlock()

	defer func() {
		close(resp_chan)
		http_api_pending_pool_mutex.Lock()
		delete(http_api_pending_pool, unique_id)
		http_api_pending_pool_mutex.Unlock()
	}()

	err = ws_push_msg(body)
	if err != nil {
		http_write_error(w, 503, "service unavailable: websocket send queue is full")
		custom_log("Error", "http server: ws push queue failed: %v", err)
		return
	}

	var timeout_ch <-chan time.Time
	if http_api_timeout > 0 {
		timer := time.NewTimer(http_api_timeout)
		defer timer.Stop()
		timeout_ch = timer.C
	}

	select {
	case response := <-resp_chan:
		w.WriteHeader(200)
		w.Write([]byte(response))
	case <-timeout_ch:
		http_write_error(w, 504, "backend ws server timeout")
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func http_write_error(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"code": code,
		"msg":  msg,
		"data": nil,
	})

	if err != nil {
		custom_log("Warn", "http server: failed to write error response: %v", err)
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}
