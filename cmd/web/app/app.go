package app

import (
	"errors"
	"fmt"
	"time"
)

// ManualControl is an interface for accepting income commands from Web Server
type ManualControl interface {
	TransferCommand(string)
	Start()
}

// RobotAccessLayer is an interface for RAL usage from Application
type RobotAccessLayer interface {
	SendCommand(string) error
	PingRobot() error
}

// Application is responsible for all logics and communicates with other layers
type Application struct {
	Robot     RobotAccessLayer
}

// TransferCommand parses command and determines what to do with it
func (app *Application) TransferCommand(command string) {
	// TO DO: parse SWAP here
}

// NewApplication constructs Application
func NewApplication() (*Application, error) {
	res := &Application{}
	return res, nil
}

func (app *Application) loop() {
	
}

// Start initializes AI process
func (app *Application) Start() {
	go app.loop()
}