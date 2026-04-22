//go:build windows

package framework

import "os"

func notifySignals() []os.Signal {
	return []os.Signal{
		os.Interrupt,
	}
}
