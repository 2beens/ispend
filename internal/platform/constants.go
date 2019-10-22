package platform

import (
	"errors"

	"github.com/2beens/ispend/internal/models"
)

var ErrNotFound = errors.New("not found")

var EmptySignal = models.Signal{}

const (
	IPAddress   = "localhost"
	DefaultPort = "8080"
)
