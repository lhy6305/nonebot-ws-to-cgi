package main

import ()

func main() {

	config_init()

	go http_server_loop()

	go ws_conn_loop()

	go ws_push_msg_loop()
	go ws_pull_msg_loop()

	select {}
}
