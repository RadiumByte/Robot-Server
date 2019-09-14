package app

import (
	"fmt"
	"math"
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
	Turn(int)
	SetSpeed(int)
	DirectCommand(string)
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
		fmt.Println("Cascade type changed to Yield Sign")
	}
}

// ProcessCommand parses command and determines what to do with it
func (app *Application) ProcessCommand(command string) {
	if command == "halt" {
		app.ChangeBlocking(true)
		app.Robot.DirectCommand("HALT")

	} else if command == "go" {
		app.ChangeBlocking(false)

	} else if command == "manual" {
		app.ChangeManual(true)
		app.Robot.DirectCommand("HALT")

	} else if command == "auto" {
		app.ChangeManual(false)
		app.Robot.DirectCommand("HALT")

	} else if command == "stopsign" {
		app.ChangeCascade(0)
		app.Robot.DirectCommand("HALT")

	} else if command == "circlesign" {
		app.ChangeCascade(1)
		app.Robot.DirectCommand("HALT")

	} else if command == "yieldsign" {
		app.ChangeCascade(2)
		app.Robot.DirectCommand("HALT")

	} else {
		// Manual control block
		firstChar := command[0]
		if firstChar == 'S' || firstChar == 'F' || firstChar == 'B' {
			if !app.IsBlocked {
				if app.IsManual {
					app.Robot.DirectCommand(command)
				}
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
	res.CascadeType = 0
	res.IsBlocked = true

	return res, nil
}

func (app *Application) ai() {
	webcam, err := gocv.OpenVideoCapture(0)
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
	imgStopSign = gocv.IMRead("stop.jpg", 0)

	imgCircleSign := gocv.NewMat()
	defer imgCircleSign.Close()
	imgCircleSign = gocv.IMRead("circle.jpg", 0)

	imgYieldSign := gocv.NewMat()
	defer imgYieldSign.Close()
	imgYieldSign = gocv.IMRead("yield.jpg", 0)

	cascadeCircle := gocv.NewCascadeClassifier()
	cascadeCircle.Load("circle.xml")

	cascadeStop := gocv.NewCascadeClassifier()
	cascadeStop.Load("stop.xml")

	cascadeTrapeze := gocv.NewCascadeClassifier()
	cascadeTrapeze.Load("yield.xml")

	blue := color.RGBA{0, 0, 255, 0}

	// "Memory" about previous correct target
	var prevTargetCenter image.Point
	var prevTargetSquare float64

	var isFirstIteration bool = true

	// Constants for object filtering
	// Don't set them here
	var MAX_DISTANCE_DIFF float64
	var MAX_SQUARE_DIFF float64
	var MAX_SIMILARITY_RATE float64
	var MIN_SIMILARITY_RATE float64

	comparator := contrib.ColorMomentHash{}

	failureCounter := 0

	app.Robot.SetSpeed(15)

	m.Lock()
	_ = webcam.Read(&imgCurrent)
	m.Unlock()

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

				// rawObjects stores all trash, which Haar Cascade returned
				var rawObjects []image.Rectangle

				if app.CascadeType == 0 {
					// Stop cascade
					rawObjects = cascadeStop.DetectMultiScale(imgCurrent)

					MAX_DISTANCE_DIFF = 400.0
					MAX_SQUARE_DIFF = 40000.0
					MAX_SIMILARITY_RATE = 100.0
					MIN_SIMILARITY_RATE = 1.0

				} else if app.CascadeType == 1 {
					// Circle cascade
					rawObjects = cascadeCircle.DetectMultiScale(imgCurrent)

					MAX_DISTANCE_DIFF = 200.0
					MAX_SQUARE_DIFF = 20000.0
					MAX_SIMILARITY_RATE = 35.0
					MIN_SIMILARITY_RATE = 1.0

				} else if app.CascadeType == 2 {
					// Yield cascade
					rawObjects = cascadeTrapeze.DetectMultiScale(imgCurrent)

					MAX_DISTANCE_DIFF = 600.0
					MAX_SQUARE_DIFF = 60000.0
					MAX_SIMILARITY_RATE = 140.0
					MIN_SIMILARITY_RATE = 1.0
				}

				if failureCounter >= 20 {
					// All normal targets disappeared, so car stoppes
					app.Robot.DirectCommand("HALT")
				}

				if len(rawObjects) == 0 {
					fmt.Println("Cascade returned empty result")
					failureCounter = failureCounter + 1
					continue
				}

				// nearObjects stores targets, which passed geometrical conditions
				var nearObjects []image.Rectangle

				// trustedObjects stores targets, which passed content and geometrical conditions
				var trustedObjects []image.Rectangle

				if isFirstIteration {
					// At first iteration simply determine biggest object
					// It is unsafe, but works only once

					// Max square calculation
					maxSquareIndex := -1
					maxSquare := 0.0
					for i, rect := range rawObjects {
						currentSquare := float64(rect.Dx() * rect.Dy())
						if currentSquare > maxSquare {
							maxSquare = currentSquare
							maxSquareIndex = i
						}
					}
					trustedObjects = append(trustedObjects, rawObjects[maxSquareIndex])

				} else {
					// Generic multi-step filtering, enables after first iteration
					// First step - find geometrically close objects by their square and location

					for _, rect := range rawObjects {
						square := rect.Dx() * rect.Dy()
						var centroidCurrent image.Point
						centroidCurrent.X = (rect.Dx() / 2) + rect.Min.X
						centroidCurrent.Y = (rect.Dy() / 2) + rect.Min.Y

						fmt.Print("Distance between centers: ")
						fmt.Println(distBetweenPoints(centroidCurrent, prevTargetCenter))
						fmt.Print("Squares difference: ")
						fmt.Println(math.Abs(float64(square) - prevTargetSquare))

						if distBetweenPoints(centroidCurrent, prevTargetCenter) < MAX_DISTANCE_DIFF {
							if math.Abs(float64(square)-prevTargetSquare) < MAX_SQUARE_DIFF {
								nearObjects = append(nearObjects, rect)
							}
						}
					}

					// Second step - usage of Color Moment Hash to compare target with preloaded model
					computedCurrent := gocv.NewMat()
					defer computedCurrent.Close()

					computedTarget := gocv.NewMat()
					defer computedTarget.Close()

					regionTarget := gocv.NewMat()
					defer regionTarget.Close()

					// Precalculated models
					if app.CascadeType == 0 {
						comparator.Compute(imgStopSign, &computedTarget)
					} else if app.CascadeType == 1 {
						comparator.Compute(imgCircleSign, &computedTarget)
					} else if app.CascadeType == 2 {
						comparator.Compute(imgYieldSign, &computedTarget)
					}

					for _, rect := range nearObjects {
						regionCurrent := gocv.NewMat()
						defer regionCurrent.Close()
						regionCurrent = imgCurrent.Region(rect)

						comparator.Compute(regionCurrent, &computedCurrent)
						similarity := comparator.Compare(computedCurrent, computedTarget)
						fmt.Print("CMH similarity: ")
						fmt.Printf("%0.4f\n", similarity)

						if similarity < MAX_SIMILARITY_RATE && similarity > MIN_SIMILARITY_RATE {
							trustedObjects = append(trustedObjects, rect)
						}
					}
				}

				if len(trustedObjects) == 0 {
					fmt.Println("No trusted objects found")
					continue
				}

				// finalObject stores only one target, selected by filter
				var finalObject image.Rectangle

				// This situation occurs if several good targets found
				// Need to find largest (closest) good target
				if len(trustedObjects) > 1 {
					maxSquareIndex := -1
					maxSquare := 0.0
					for i, rect := range trustedObjects {
						currentSquare := float64(rect.Dx() * rect.Dy())
						if currentSquare > maxSquare {
							maxSquare = currentSquare
							maxSquareIndex = i
						}
					}
					finalObject = trustedObjects[maxSquareIndex]
				}

				// This counter checks how many times cascade returned empty data
				failureCounter = 0

				// Car throttle logic
				// Throttle depends on a distance to target
				// Bigger square - lower throttle down to full stop

				// Car behaviour configuration
				// Change car's speed while driving forward
				var maxThrottle int = 60
				var minThrottle int = 15

				// Change distances for min and max speed
				// Min speed:
				var maxSquare float64 = 700 * 700
				// Max speed:
				var minSquare float64 = 100 * 100

				// How fast car will accelerate backward
				var backwardAcceleration float64 = 20000.0

				// Max speed for driving backward
				var maxBackwardThrottle = 50

				// Actual size of the target
				targetSquare := float64(finalObject.Dx() * finalObject.Dy())

				if targetSquare > maxSquare {
					// Target is too close - car is going backward

					deltaSquare := targetSquare - maxSquare
					calculatedThrottle := int(deltaSquare / backwardAcceleration)
					calculatedThrottleStr := strconv.Itoa(calculatedThrottle)

					app.Robot.DirectCommand("B" + calculatedThrottleStr)
				} else {
					// Target in range - car is going forward

					targetSquareInRange := maxSquare - targetSquare
					deltaThrottle := maxThrottle - minThrottle
					deltaSquare := math.Abs(maxSquare - minSquare)

					calculatedThrottle := int(((float64(deltaThrottle) * targetSquareInRange) / deltaSquare) + float64(minThrottle))
					app.Robot.SetSpeed(calculatedThrottle)
				}

				// Car steering logic
				// Horizontal position of target influences on wheels steering

				var command int

				// Calculate center of the target
				var centroid image.Point
				centroid.X = (finalObject.Dx() / 2) + finalObject.Min.X
				centroid.Y = (finalObject.Dy() / 2) + finalObject.Min.Y

				// Refresh previous center and square
				prevTargetCenter = centroid
				prevTargetSquare = float64(finalObject.Dx() * finalObject.Dy())

				isFirstIteration = false

				// Borders determine tube area in the center of frame
				// In this area car won't steer in any directions
				rightBorder := int(float64(imgCurrent.Cols()) * 0.53)
				leftBorder := int(float64(imgCurrent.Cols()) * 0.47)

				if centroid.X >= leftBorder && centroid.X <= rightBorder {
					// Taeget in tube - need to ride forward
					command = 50

				} else if centroid.X < leftBorder {
					// Need to steer left
					command = (50 * centroid.X) / leftBorder

				} else if centroid.X > rightBorder {
					// Need to steer right
					command = ((50 * centroid.X) / (imgCurrent.Cols() - rightBorder)) - 12
				}

				// Max turning if target is going to escape frame
				// In theory, this will increase steering ability
				if command >= 80 {
					command = 100
				} else if command <= 20 {
					command = 0
				}

				app.Robot.Turn(command)

				// Draw bounding box and show it
				gocv.Rectangle(&imgCurrent, finalObject, blue, 3)
				size := gocv.GetTextSize("Target", gocv.FontHersheyPlain, 1.2, 2)
				pt := image.Pt(finalObject.Min.X+(finalObject.Min.X/2)-(size.X/2), finalObject.Min.Y-2)
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
