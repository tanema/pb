package util

import (
	"os"
	"os/signal"
)

func OnSignal(fn func(os.Signal), sig ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, sig...)
	for {
		fn(<-c)
	}
}
