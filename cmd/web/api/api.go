package api

import (
	"fmt"
	"strconv"

	"github.com/DiaElectronics/online_kasse/cmd/web/app"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

type WebServer struct {
	application app.ManualControl
}

// PushCommand pushes new Command to Application
func (server *WebServer) PushCommand(ctx *fasthttp.RequestCtx) {
	commandStr := ctx.UserValue("command").(string)
	server.application.TransferCommand(commandStr)
}

// Start initializes Web Server, starts application and begins serving
func (server *WebServer) Start() {
	server.application.Start()

	router := fasthttprouter.New()
	router.PUT("/:command", server.PushCommand)

	port := ":8080"

	fmt.Println("Server is starting on port", port)
	fasthttp.ListenAndServe(port, router.Handler)
}

// NewWebServer constructs Web Server
func NewWebServer(application app.IncomeRegistration) (*WebServer, error) {
	res := &WebServer{}
	res.application = application

	return res, nil
}

