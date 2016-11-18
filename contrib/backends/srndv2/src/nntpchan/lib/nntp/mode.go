package nntp

import (
	"errors"
	"strings"
)

var ErrInvalidMode = errors.New("invalid mode set")

// a mode set by an nntp client
type Mode string

// reader mode
const MODE_READER = Mode("reader")

// streaming mode
const MODE_STREAM = Mode("stream")

// mode is not set
const MODE_UNSET = Mode("")

// get as string
func (m Mode) String() string {
	return strings.ToUpper(string(m))
}

// is this a valid mode of operation?
func (m Mode) Valid() bool {
	return m.Is(MODE_READER) || m.Is(MODE_STREAM)
}

// is this mode equal to another mode
func (m Mode) Is(other Mode) bool {
	return m.String() == other.String()
}

// a switch mode command
type ModeCommand string

// get as string
func (m ModeCommand) String() string {
	return strings.ToUpper(string(m))
}

// is this mode command well formed?
// does not check the actual mode sent.
func (m ModeCommand) Valid() bool {
	s := m.String()
	return strings.Count(s, " ") == 1 && strings.HasPrefix(s, "MODE ")
}

// get the mode selected in this mode command
func (m ModeCommand) Mode() Mode {
	return Mode(strings.Split(m.String(), " ")[1])
}

// check if this mode command is equal to an existing one
func (m ModeCommand) Is(cmd ModeCommand) bool {
	return m.String() == cmd.String()
}

// reader mode command
const ModeReader = ModeCommand("mode reader")

// streaming mode command
const ModeStream = ModeCommand("mode stream")

// line prefix for mode
const LinePrefix_Mode = "MODE "
