package shutdown

import (
	"os"
	"os/signal"
	"syscall"
)

// WaitForShutdownSignal waits for OS shutdown signals (SIGINT, SIGTERM)
// Returns the received signal
func WaitForShutdownSignal() os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	return <-sigChan
}

