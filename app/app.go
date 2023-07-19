package app

import (
	"io"
	"os"
)

type App struct {
	Out io.Writer
}

func New() (*App, error) {
	return &App{
		Out: os.Stdout,
	}, nil
}

