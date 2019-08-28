package api

import (
	"fmt"
	//"strconv"

	"github.com/RadiumByte/Robot-Server/cmd/web/app"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

type WebServer struct {
	application app.ManualControl
}

// PushCommand pushes new Command to Application
func (server *WebServer) PushCommand(ctx *fasthttp.RequestCtx) {
	fmt.Println("Controller started")
	commandStr := ctx.UserValue("command").(string)
	fmt.Println("API got data: " + commandStr)
	server.application.TransferCommand(commandStr)
}

// Start initializes Web Server, starts application and begins serving
func (server *WebServer) Start(port string) {
	server.application.Start()

	router := fasthttprouter.New()
	router.POST("/:command", server.PushCommand)

	port = ":8080"
	fmt.Println("Server is starting on port" + port)
	fmt.Println(fasthttp.ListenAndServe(port, router.Handler))
}

// NewWebServer constructs Web Server
func NewWebServer(application app.ManualControl) (*WebServer, error) {
	res := &WebServer{}
	res.application = application

	return res, nil
}

