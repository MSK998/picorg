package app

import (
	"io"
	"os"

	"github.com/MSK998/picorg/internal/fswrapper"
)

type App struct {
	Out io.Writer
	Fs fswrapper.FS
}

func New(fs fswrapper.FS) (*App, error) {
	return &App{
		Out: os.Stdout,
		Fs: fs,
	}, nil
}

