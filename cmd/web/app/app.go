package app

import (
	"fmt"
	"sync"

	"image"
	"image/color"

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
	Robot    RobotAccessLayer
	IsManual bool
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
	fmt.Println("Mode changed")
}

// TransferCommand parses command and determines what to do with it
func (app *Application) TransferCommand(command string) {
	if command == "swap" {
		app.ChangeMode()
	} else {
		firstChar := command[0]
		if firstChar == 'l' || firstChar == 'r' || firstChar == 'f' || firstChar == 'b' {
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

	imgResult := gocv.NewMat()
	defer imgResult.Close()

	mask := gocv.NewMat()
	defer mask.Close()

	//var imageSize int = 200
	//var matchThreshold int = 3

	cascade := gocv.NewCascadeClassifier()
	cascade.Load("cups.xml")

	imgCurrent = gocv.IMRead("sign2.jpg", 0)

	fmt.Printf("Main loop is starting...")
	for {
		if !app.IsManual {
			//m.Lock()
			//ok := webcam.Read(&imgCurrent)
			//m.Unlock()

			//if !ok {
			//	fmt.Printf("Error while read RTSP: program aborted...")
			//	return
			//}
			if imgCurrent.Empty() {
				continue
			}

			var min image.Point
			min.X = 0
			min.Y = 0
			var max image.Point
			max.X = 0
			max.Y = 0

			var col color.RGBA
			col.B = 255

			targets := cascade.DetectMultiScale(imgCurrent)
			//targets := cascade.DetectMultiScaleWithParams(imgCurrent, 1.4, 3, 0, min, max)
			if len(targets) == 0 {
				fmt.Println("No objects found")
				continue
			}
			first := targets[0]

			gocv.Rectangle(&imgCurrent, first, col, 2)

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
