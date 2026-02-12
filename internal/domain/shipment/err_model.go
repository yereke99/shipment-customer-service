package shipment

import (
	"errors"
	"regexp"
)

var idnPattern = regexp.MustCompile(`^\d{12}$`)

var (
	ErrInvalidRoute      = errors.New("invalid route")
	ErrInvalidPrice      = errors.New("invalid price")
	ErrInvalidIDN        = errors.New("invalid idn")
	ErrInvalidShipmentID = errors.New("invalid shipment id")
	ErrNotFound          = errors.New("shipment not found")
)

func IsValidIDN(value string) bool {
	return idnPattern.MatchString(value)
}
