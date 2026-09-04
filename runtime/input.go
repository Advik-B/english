package runtime

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"
)

// Reading a line of input goes through here, and there is one reader.
//
// There were three, each built fresh for every read: one in the tree-walking
// engine, one in the instruction engine and one in the standard library's
// "ask". A buffered reader reads ahead — as much as is available, not as much
// as was asked for — and each of these discarded its buffer when it went out
// of scope, so two consecutive questions lost whatever had been typed past the
// first line. Two questions answered by a pipe, or by someone typing quickly,
// would leave the second holding nothing.
//
// One reader for the process keeps what it read ahead, so the next question
// gets it.
var (
	inputOnce   sync.Once
	inputReader *bufio.Reader
)

// SetInput replaces the source of input, for a test or an embedding program.
// It resets what has been read ahead, which is the point.
func SetInput(r io.Reader) {
	inputOnce.Do(func() {})
	inputReader = bufio.NewReader(r)
}

// ReadLine reads one line of input, without its line ending.
//
// It returns the empty string at end of input, which is what a program that
// asks a question of an exhausted pipe should see.
func ReadLine() string {
	inputOnce.Do(func() {
		if inputReader == nil {
			inputReader = bufio.NewReader(os.Stdin)
		}
	})

	line, err := inputReader.ReadString('\n')
	if err != nil && line == "" {
		return ""
	}
	return strings.TrimRight(line, "\r\n")
}
