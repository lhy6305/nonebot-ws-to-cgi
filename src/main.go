package main

import (
	"os"
)

func main() {
	custom_log("Info", "Program started")
	defer custom_log("Info", "Program exited")

	if instance_pid := process_get_another_instance_pid(); instance_pid > 0 {
		if !allow_multi_instance {
			custom_log("Fatal", "A same instance (pid %d) is already running", instance_pid)
			os.Exit(1)
			return
		}
		custom_log("Warn", "")
		custom_log("Warn", "A same instance (pid %d) is already running", instance_pid)
		custom_log("Warn", "")
	}

	config_init()

	go http_server_loop()

	go ws_conn_loop()

	go ws_push_msg_loop()
	go ws_pull_msg_loop()

	select {}
}
