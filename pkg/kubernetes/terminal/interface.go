package terminal

import (
	"io"
	"k8s.io/client-go/tools/remotecommand"
)

const EndOfTransmission = "\u0004"

type Handler interface {
	io.Reader
	io.Writer
	remotecommand.TerminalSizeQueue
}
