package app

import (
	"fmt"
	"sync"

	//"image"
	//"image/color"

	"gocv.io/x/gocv"
)

// bufferEraser cleans input videostream from unnecessary frames
func bufferEraser(source *gocv.VideoCapture, m *sync.Mutex) {
	tmp := gocv.NewMat()
	defer tmp.Close()

	for {
		m.Lock()
		_ = source.Read(&tmp)
		m.Unlock()
	}
}

// RobotServer is an interface for accepting income commands from Web Server
type RobotServer interface {
	ProcessCommand(string)
	ChangeMode()
	ChangeCascade(int8)
	Start()
}

// RobotAccessLayer is an interface for RAL usage from Application
type RobotAccessLayer interface {
	SendCommand(string)
}

// Application is responsible for all logics and communicates with other layers
type Application struct {
	Robot       RobotAccessLayer
	IsManual    bool
	CascadeType int8
}

// ChangeMode swaps current mode of application
func (app *Application) ChangeMode() {
	app.IsManual = !app.IsManual
	fmt.Println("Driving mode changed")
}

// ChangeCascade changes cascade, assigned to the specific sign
// 0 - stop
// 1 - circle
// 2 - trapeze
func (app *Application) ChangeCascade(cascade int8) {
	app.CascadeType = cascade
	if cascade == 0 {
		fmt.Println("Cascade type changed to Stop Sign")
	} else if cascade == 1 {
		fmt.Println("Cascade type changed to Circle Sign")
	} else if cascade == 2 {
		fmt.Println("Cascade type changed to Trapeze Sign")
	}
}

// ProcessCommand parses command and determines what to do with it
func (app *Application) ProcessCommand(command string) {
	if command == "swap" {
		app.ChangeMode()
	} else if command == "stop" {
		app.ChangeCascade(0)
		fmt.Println("Stop set")
	} else if command == "circle" {
		app.ChangeCascade(1)
		fmt.Println("Circle set")
	} else if command == "trapeze" {
		app.ChangeCascade(2)
		fmt.Println("Trapeze set")
	} else {
		firstChar := command[0]
		if firstChar == 's' || firstChar == 'f' || firstChar == 'b' {
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
	webcam, err := gocv.OpenVideoCapture("rtsp://81.23.197.208/user=admin_password=8555_channel=0_stream=0.sdp")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer webcam.Close()
	fmt.Println("RTSP videostream claimed...")

	var m sync.Mutex
	go bufferEraser(webcam, &m)
	fmt.Println("Buffer eraser started...")

	window := gocv.NewWindow("Autopilot")
	defer window.Close()

	imgCurrent := gocv.NewMat()
	defer imgCurrent.Close()

	/*
		cascadeCircle := gocv.NewCascadeClassifier()
		cascadeCircle.Load("circle.xml")

		cascadeStop := gocv.NewCascadeClassifier()
		cascadeStop.Load("stop.xml")

		cascadeTrapeze := gocv.NewCascadeClassifier()
		cascadeTrapeze.Load("trapeze.xml")
	*/

	fmt.Printf("Main loop is starting...")
	for {
		if !app.IsManual {
			m.Lock()
			ok := webcam.Read(&imgCurrent)
			m.Unlock()

			if !ok {
				fmt.Printf("Error while read RTSP: program aborted...")
				return
			}
			if imgCurrent.Empty() {
				continue
			}

			//target := cascade.DetectMultiScale(imgCurrent)

			// TO DO: make image processing here
			// TO DO: make car driving here

			window.IMShow(imgCurrent)
			if window.WaitKey(1) >= 0 {
				break
			}
		}
	}
}

// Start initializes AI process
func (app *Application) Start() {
	go app.ai()
}
