package app

import (
	//"errors"
	"fmt"
)

// ManualControl is an interface for accepting income commands from Web Server
type ManualControl interface {
	TransferCommand(string)
	ChangeMode()
	Start()
}

// RobotAccessLayer is an interface for RAL usage from Application
type RobotAccessLayer interface {
	SendCommand(string)
}

// Application is responsible for all logics and communicates with other layers
type Application struct {
	Robot     RobotAccessLayer
	IsManual  bool
}

			/*
			valueStr := command[1:len(command)]
			value, err := strconv.ParseInt(valueStr)

			if err != nil {
				fmt.Fprintf(ctx, "Invalid type of parameter: might be int\n")
				return
			}
			*/

// ChangeMode swaps current mode of application
func (app *Application) ChangeMode() {
	app.IsManual = !app.IsManual
}

// TransferCommand parses command and determines what to do with it
func (app *Application) TransferCommand(command string) {
	if command == "swap" {
		app.ChangeMode()
	} else {
		firstChar := command[0]
		if firstChar == 'L' || firstChar == 'R' || firstChar == 'F' || firstChar == 'B' {
			fmt.Println("Application verified data: " + command)
			app.Robot.SendCommand(command)
		}
	}
}

// NewApplication constructs Application
func NewApplication(robot RobotAccessLayer) (*Application, error) {
	res := &Application{}
	res.Robot = robot
	
	return res, nil
}

func (app *Application) ai() {

	for {
		if !app.IsManual {

		}
	}
}

// Start initializes AI process
func (app *Application) Start() {
	go app.ai()
}