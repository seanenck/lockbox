// Package platform handles platform-specific operations.
package platform

import (
	"errors"

	"github.com/enckse/lockbox/internal/config"
)

const (
	termBeginRed = "\033[1;31m"
	termEndRed   = "\033[0m"
)

const (
	// Red will get red terminal coloring.
	Red = iota
)

type (
	// Color are terminal colors for dumb terminal coloring.
	Color int
)

type (
	// Terminal represents terminal coloring information.
	Terminal struct {
		Start string
		End   string
	}
)

// NewTerminal will retrieve start/end terminal coloration indicators.
func NewTerminal(color Color) (Terminal, error) {
	if color != Red {
		return Terminal{}, errors.New("bad color")
	}
	interactive, err := config.EnvInteractive.Get()
	if err != nil {
		return Terminal{}, err
	}
	colors := interactive
	if colors {
		isColored, err := config.EnvNoColor.Get()
		if err != nil {
			return Terminal{}, err
		}
		colors = !isColored
	}
	if colors {
		return Terminal{Start: termBeginRed, End: termEndRed}, nil
	}
	return Terminal{}, nil
}