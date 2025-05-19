package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	ws_conn             *websocket.Conn = nil
	ws_pull_buffer_chan                 = make(chan []byte, ws_max_pull_queue_size)
	ws_push_buffer_chan                 = make(chan []byte, ws_max_push_queue_size)
)

// async call
func ws_conn_loop() {
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		},
	}

	var err error = nil

	for {
		if ws_conn != nil {
			ws_conn.Close()
		}

		custom_log("Info", "Connecting to WebSocket server")

		header := http.Header{}
		header.Set("Authorization", ws_auth_token)

		ws_conn, _, err = dialer.Dial(wss_url, header)
		if err != nil {
			custom_log("Error", "WebSocket dial error: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		custom_log("Info", "Connected to WebSocket server")

		handle_ws_conn() // will block here until error

		custom_log("Info", "Connection closed, reconnecting...")
		time.Sleep(500 * time.Millisecond)
		continue
	}
}

// sync call: block until error
func handle_ws_conn() {
	for {
		msg_type, msg, err := ws_conn.ReadMessage()
		if err != nil {
			custom_log("Error", "WebSocket recv error: %v", err)
			return
		}

		custom_log("Debug", "WebSocket recv %d bytes, type %v", len(msg), msg_type)
		custom_log("Trace", "%v", string(msg))

		if msg_type != websocket.TextMessage && msg_type != websocket.BinaryMessage {
			custom_log("Warn", "WebSocket recv %d bytes, type %v", len(msg), msg_type)
			custom_log("Debug", "%v", string(msg))
			continue
		}
		var data map[string]interface{}
		if err := json.Unmarshal(msg, &data); err != nil {
			custom_log("Error", "ws_conn: invalid JSON: %v", err)
			continue
		}

		select {
		case ws_pull_buffer_chan <- msg:
		// msg push succ
		default:
			custom_log("Warn", "ws message queue overflow")
		}
	}
}

// async call
func ws_pull_msg_loop() {
	for {
		msg := <-ws_pull_buffer_chan

		var msg_json_map map[string]interface{}
		if err := json.Unmarshal(msg, &msg_json_map); err != nil {
			custom_log("Error", "invalid JSON from WebSocket: %v", err)
			custom_log("Debug", "%v", string(msg))
			continue
		}

		unique_id_t, ok := msg_json_map["echo"]
		if !ok {
			go call_cgi_wrapper(msg)
			continue
		}

		delete(msg_json_map, "echo")

		unique_id, ok := unique_id_t.(string)
		if !ok {
			custom_log("Warn", "the type of unique request id is not string. value: %v", unique_id_t)
			go call_cgi_wrapper(msg)
			continue
		}

		http_api_pending_pool_mutex.Lock()
		ch, exists := http_api_pending_pool[unique_id]
		if !exists {
			custom_log("Warn", "unique request id %s is not found in pool", unique_id)
			go call_cgi_wrapper(msg)
			continue
		}
		delete(http_api_pending_pool, unique_id)
		http_api_pending_pool_mutex.Unlock()

		resp, err := json.Marshal(msg_json_map)
		if err != nil {
			ch <- []byte{}
			custom_log("Warn", "failed to marshal json: %v", err)
			go call_cgi_wrapper(msg)
			continue
		}

		ch <- resp
	}
}

// async call
func ws_push_msg_loop() {
	for data := range ws_push_buffer_chan {
		custom_log("Debug", "WebSocket send %d bytes", len(data))
		custom_log("Trace", "%v", string(data))
		for {
			err := ws_conn.WriteMessage(websocket.TextMessage, data)
			if err == nil {
				break
			}
			custom_log("Error", "write to ws failed, retrying after 1 second...")
			time.Sleep(1 * time.Second)
		}
	}
}

func ws_push_msg(data []byte) error {
	select {
	case ws_push_buffer_chan <- data:
		// msg push succ
		return nil
	default:
		break
	}
	return fmt.Errorf("ws message queue overflow")
}
