package app

import (
	"fmt"
	"math"
	"strconv"
	"sync"

	"image"
	"image/color"

	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
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

	ChangeBlocking(bool)
	ChangeManual(bool)

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
	IsBlocked   bool
	CascadeType int8
}

// ChangeBlocking can block/unblock car movements
func (app *Application) ChangeBlocking(mode bool) {
	app.IsBlocked = mode

	if mode {
		fmt.Println("Car is blocked")
	} else {
		fmt.Println("Car is moving")
	}
}

// ChangeBlocking sets current mode of driving
func (app *Application) ChangeManual(mode bool) {
	app.IsManual = mode

	if mode {
		fmt.Println("Car is on manual control")
	} else {
		fmt.Println("Car is driving automatically")
	}
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
	if command == "halt" {
		app.ChangeBlocking(true)

	} else if command == "go" {
		app.ChangeBlocking(false)

	} else if command == "manual" {
		app.ChangeManual(true)

	} else if command == "auto" {
		app.ChangeManual(false)

	} else if command == "stopsign" {
		app.ChangeCascade(0)

	} else if command == "circlesign" {
		app.ChangeCascade(1)

	} else if command == "trapezesign" {
		app.ChangeCascade(2)

	} else {
		firstChar := command[0]
		if firstChar == 's' || firstChar == 'f' || firstChar == 'b' {
			if !app.IsBlocked {
				app.Robot.SendCommand(command)
			}
		}
	}
}

func distBetweenPoints(from image.Point, to image.Point) float64 {
	return math.Sqrt(float64((to.X-from.X)*(to.X-from.X) + (to.Y-from.Y)*(to.Y-from.Y)))
}

// NewApplication constructs Application
func NewApplication(robot RobotAccessLayer) (*Application, error) {
	res := &Application{}
	res.Robot = robot
	res.CascadeType = 2
	return res, nil
}

func (app *Application) ai() {
	webcam, err := gocv.OpenVideoCapture("rtsp://192.168.1.39:8080/video/h264")
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

	imgTarget := gocv.NewMat()
	defer imgTarget.Close()

	imgStopSign := gocv.NewMat()
	defer imgStopSign.Close()
	imgStopSign = gocv.IMRead("yield.jpg", 0)

	cascadeCircle := gocv.NewCascadeClassifier()
	cascadeCircle.Load("circle.xml")

	cascadeStop := gocv.NewCascadeClassifier()
	cascadeStop.Load("stop.xml")

	cascadeTrapeze := gocv.NewCascadeClassifier()
	cascadeTrapeze.Load("yield.xml")

	blue := color.RGBA{0, 0, 255, 0}

	// "Memory" about target
	var prevTargetCenter image.Point
	var prevTargetSquare float64
	var prevTargetRect image.Rectangle

	var isFirstIteration bool = true

	// Use these constants to change object filtering
	var MAX_DISTANCE_DIFF float64 = 500.0
	var MAX_SQUARE_DIFF float64 = 15000.0
	var MIN_SIMILARITY_RATE float64 = 60.0

	comparator := contrib.ColorMomentHash{}

	fmt.Println("Main loop is starting...")
	for {
		if !app.IsManual {
			if !app.IsBlocked {
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

				fmt.Println("Image got")

				var target []image.Rectangle

				if app.CascadeType == 0 {
					// Stop cascade
					target = cascadeStop.DetectMultiScale(imgCurrent)
					fmt.Println("Using target cascade")

				} else if app.CascadeType == 1 {
					// Circle cascade
					target = cascadeCircle.DetectMultiScale(imgCurrent)

				} else if app.CascadeType == 2 {
					// Trapeze cascade
					target = cascadeTrapeze.DetectMultiScale(imgCurrent)
				}

				if len(target) == 0 {
					fmt.Println("Cascade returned empty result")
					continue
				}

				var closeTargets []image.Rectangle
				var trustedObjects []image.Rectangle

				if isFirstIteration {
					// At first iteration simply determine biggest object
					fmt.Println("First iteration - find max square")

					maxSquareIndex := -1
					maxSquare := 0.0
					for i, r := range target {
						currentSquare := float64(r.Dx() * r.Dy())
						if currentSquare > maxSquare {
							maxSquare = currentSquare
							maxSquareIndex = i
							fmt.Print("Max square updated to ")
							fmt.Println(maxSquare)
						}
					}
					trustedObjects = append(trustedObjects, target[maxSquareIndex])

				} else {
					// First step of filtering - find geometrically close objects
					fmt.Println("Generic iteration - filtering")

					for _, r := range target {
						square := r.Dx() * r.Dy()
						var centroidCurrent image.Point
						centroidCurrent.X = (r.Dx() / 2) + r.Min.X
						centroidCurrent.Y = (r.Dy() / 2) + r.Min.Y

						fmt.Print("Distance between centers: ")
						fmt.Println(distBetweenPoints(centroidCurrent, prevTargetCenter))
						fmt.Print("Squares difference: ")
						fmt.Println(math.Abs(float64(square) - prevTargetSquare))

						if distBetweenPoints(centroidCurrent, prevTargetCenter) < MAX_DISTANCE_DIFF {
							if math.Abs(float64(square)-prevTargetSquare) < MAX_SQUARE_DIFF {
								closeTargets = append(closeTargets, r)
							}
						}
					}

					computedCurrent := gocv.NewMat()
					defer computedCurrent.Close()

					computedTarget := gocv.NewMat()
					defer computedTarget.Close()

					regionTarget := gocv.NewMat()
					defer regionTarget.Close()
					//regionTarget = imgCurrent.Region(prevTargetRect)
					comparator.Compute(imgStopSign, &computedTarget)

					// Second step of filtering - determine object with high similarity rate
					for _, r := range closeTargets {
						regionCurrent := gocv.NewMat()
						defer regionCurrent.Close()
						regionCurrent = imgCurrent.Region(r)

						comparator.Compute(regionCurrent, &computedCurrent)
						similarity := comparator.Compare(computedCurrent, computedTarget)
						fmt.Print("CMH similarity: ")
						fmt.Printf("%0.4f\n", similarity)

						if similarity < MIN_SIMILARITY_RATE && similarity > 1.0 {
							trustedObjects = append(trustedObjects, r)
						}
					}
				}

				if len(trustedObjects) == 0 {
					fmt.Println("No trusted objects found")
					continue
				}

				var command string

				var centroid image.Point
				centroid.X = (trustedObjects[0].Dx() / 2) + trustedObjects[0].Min.X
				centroid.Y = (trustedObjects[0].Dy() / 2) + trustedObjects[0].Min.Y

				prevTargetCenter = centroid
				fmt.Println("Target centroid updated: " + centroid.String())
				prevTargetSquare = float64(trustedObjects[0].Dx() * trustedObjects[0].Dy())
				fmt.Print("Target square updated: ")
				fmt.Println(prevTargetSquare)

				if isFirstIteration {
					prevTargetRect = trustedObjects[0]
					fmt.Println("Target rectangle updated: " + prevTargetRect.String())

					isFirstIteration = false
				}

				rightBorder := int(float64(imgCurrent.Cols()) * 0.6)
				leftBorder := int(float64(imgCurrent.Cols()) * 0.4)

				if centroid.X >= leftBorder && centroid.X <= rightBorder {
					// Need to ride forward
					command = "S50A"

				} else if centroid.X < leftBorder {
					// Need to steer left
					var steerValue int
					steerValue = (50 * centroid.X) / leftBorder
					steerValueStr := strconv.Itoa(steerValue)
					command = "S" + steerValueStr + "A"

				} else if centroid.X > rightBorder {
					// Need to steer right
					var steerValue int
					steerValue = 100 - ((50 * centroid.X) / (imgCurrent.Cols() - rightBorder))
					steerValueStr := strconv.Itoa(steerValue)
					command = "S" + steerValueStr + "A"
				}

				app.Robot.SendCommand(command)

				gocv.Rectangle(&imgCurrent, trustedObjects[0], blue, 3)

				size := gocv.GetTextSize("Target", gocv.FontHersheyPlain, 1.2, 2)
				pt := image.Pt(trustedObjects[0].Min.X+(trustedObjects[0].Min.X/2)-(size.X/2), trustedObjects[0].Min.Y-2)
				gocv.PutText(&imgCurrent, "Target", pt, gocv.FontHersheyPlain, 1.2, blue, 2)
			}

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
