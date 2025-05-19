package main

import (
	"os"
)

func main() {

	config_init()

	if instance_pid := process_get_another_instance_pid(); instance_pid > 0 {
		custom_log("Fatal", "A same instance (pid %d) is already running", instance_pid)
		os.Exit(1)
	}

	go http_server_loop()

	go ws_conn_loop()

	go ws_push_msg_loop()
	go ws_pull_msg_loop()

	select {}
}
