package main

import (
	"fmt"
	"sync"

	"github.com/RadiumByte/Robot-Server/cmd/web/api"
	"github.com/RadiumByte/Robot-Server/cmd/web/app"
)

func main() {
	application, err := app.NewApplication()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Application start failure - program stopped")
		return
	}

	server, err := api.NewWebServer(application)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Server start failure - program stopped")
		return
	}
	server.Start()
}