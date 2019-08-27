package ral

import (
	"fmt"

	"github.com/DiaElectronics/online_kasse/cmd/web/app"
)

// RoboCar represents Raspberry Pi based car
type RoboCar struct {
	// TO DO: add fields
}

// NewRoboCar constructs object of RoboCar
func NewRoboCar() (*RoboCar, error) {
	res := &PostgresDAL{}
	return res, nil
}