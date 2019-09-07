package ral

import (
	"fmt"
	"strconv"
	//"github.com/valyala/fasthttp"
)

// RoboCar represents Raspberry Pi based car
type RoboCar struct {
	CarIP         string
	CarPort       string
	MovementSpeed int
}

// Turn creates HTTP client and sends coommand to robot
func (robot *RoboCar) Turn(steerValue int) {
	steerValueStr := strconv.Itoa(steerValue)
	command := "S" + steerValueStr + "A"
	fmt.Println("Sending command: " + command)

	//url := "http://" + robot.CarIP + robot.CarPort + "/" + command
	/*
		req := fasthttp.AcquireRequest()
		req.SetRequestURI(url)
		req.Header.SetMethod("PUT")

		resp := fasthttp.AcquireResponse()
		client := &fasthttp.Client{}
		client.Do(req, resp)

		fmt.Println("Command sent to robot: " + command)
	*/
}

// Turn creates HTTP client and sends coommand to robot
func (robot *RoboCar) DirectCommand(command string) {
	fmt.Println("Sending command: " + command)

	//url := "http://" + robot.CarIP + robot.CarPort + "/" + command
	/*
		req := fasthttp.AcquireRequest()
		req.SetRequestURI(url)
		req.Header.SetMethod("PUT")

		resp := fasthttp.AcquireResponse()
		client := &fasthttp.Client{}
		client.Do(req, resp)

		fmt.Println("Command sent to robot: " + command)
	*/
}

func (robot *RoboCar) SetSpeed(speed int) {
	robot.MovementSpeed = speed
}

// NewRoboCar constructs object of RoboCar
func NewRoboCar(ip string, port string) (*RoboCar, error) {
	res := &RoboCar{}

	res.CarPort = port
	res.CarIP = ip
	return res, nil
}
