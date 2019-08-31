package api

import (
	"fmt"

	"github.com/RadiumByte/Robot-Server/cmd/web/app"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

// WebServer is
type WebServer struct {
	application app.ManualControl
}

// PushCommand pushes new Command to Application
func (server *WebServer) PushCommand(ctx *fasthttp.RequestCtx) {
	commandStr := ctx.UserValue("command").(string)
	fmt.Println("Server received command: " + commandStr)
	server.application.TransferCommand(commandStr)
}

// Start initializes Web Server, starts application and begins serving
func (server *WebServer) Start(port string) {
	server.application.Start()

	router := fasthttprouter.New()
	router.PUT("/:command", server.PushCommand)

	fmt.Println("Server is starting on port" + port)
	fmt.Println(fasthttp.ListenAndServe(port, router.Handler))
}

// NewWebServer constructs Web Server
func NewWebServer(application app.ManualControl) (*WebServer, error) {
	res := &WebServer{}
	res.application = application

	return res, nil
}
