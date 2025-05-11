package models

type CarType string

const (
	Sedan    CarType = "SEDAN"
	SUV      CarType = "SUV"
	Minivan  CarType = "MINIVAN"
	Roadster CarType = "ROADSTER"
)

type Car struct {
	ID                 int64
	CarUID             string
	Brand              string
	Model              string
	RegistrationNumber string
	Power              uint64
	Price              uint64
	Type               CarType
	Availability       bool
}
