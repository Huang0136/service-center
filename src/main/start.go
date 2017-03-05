// bootstrap
package main

import (
	"constants"
	"handlers"
	"log"
	"monitor"
	"net/http"
	"time"
)

func main() {
	log.Println("[Server Center] project start:", time.Now().Format("2006-01-02 15:04:05.9999"))

	http.HandleFunc("/", handlers.GenericHandler)

	err := http.ListenAndServe(":"+constants.Configs["serverCenter.http.port"], nil)
	if err != nil {
		log.Fatalln("start error:", err)
	}
}

func init() {
	constants.InitConstants()

	monitor.InitMonitorEtcd()

}
