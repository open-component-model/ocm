package misc

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/mandelsoft/logging"

	"ocm.software/ocm/api/utils"
)

type Flusher interface {
	Flush() error
}

func Flush(o interface{}) error {
	if f, ok := o.(Flusher); ok {
		return f.Flush()
	}
	return nil
}

var (
	StdoutPrinter = NewPrinter(os.Stdout)
	StderrPrinter = NewPrinter(os.Stderr)
	NonePrinter   = NewPrinter(nil)
)

type Printer interface {
	io.Writer
	Printf(msg string, args ...interface{}) (int, error)

	AddGap(gap string) Printer
}

type FlushingPrinter interface {
	Printer
	Flusher
}

type printerState struct {
	mu      sync.Mutex
	pending bool
}

type printer struct {
	writer io.Writer
	gap    string
	state  *printerState // shared across derived printers
}

func NewPrinter(writer io.Writer) Printer {
	return &printer{
		writer: writer,
		state: &printerState{
			pending: true,
		},
	}
}

func AssurePrinter(p Printer) Printer {
	return utils.OptionalDefaulted(NonePrinter, p)
}

func NewBufferedPrinter() (Printer, *bytes.Buffer) {
	buf := bytes.NewBuffer(nil)
	return NewPrinter(buf), buf
}

func (p *printer) AddGap(gap string) Printer {
	return &printer{
		writer: p.writer,
		gap:    p.gap + gap,
		state:  p.state, // intentionally shared
	}
}

func (p *printer) Write(data []byte) (int, error) {
	if p.writer == nil {
		return 0, nil
	}

	p.state.mu.Lock()
	defer p.state.mu.Unlock()

	s := strings.ReplaceAll(string(data), "\n", "\n"+p.gap)
	if strings.HasSuffix(s, "\n"+p.gap) {
		p.state.pending = true
		s = s[:len(s)-len(p.gap)]
	}

	return p.writer.Write([]byte(s))
}

func (p *printer) Printf(msg string, args ...interface{}) (int, error) {
	if p == nil || p.writer == nil {
		return 0, nil
	}

	p.state.mu.Lock()
	defer p.state.mu.Unlock()

	data := fmt.Sprintf(msg, args...)

	// Prepend gap if starting a new line.
	if p.gap != "" && p.state.pending {
		data = p.gap + data
	}

	// pending = “next write starts a new line”
	if strings.HasSuffix(data, "\n") {
		p.state.pending = true
	} else {
		p.state.pending = false
	}

	return p.writer.Write([]byte(data))
}

////////////////////////////////////////////////////////////////////////////////

type loggingPrinter struct {
	log logging.Logger
	gap string

	mu      sync.Mutex
	pending string
}

func NewLoggingPrinter(log logging.Logger) FlushingPrinter {
	return &loggingPrinter{log: log}
}

func (p *loggingPrinter) AddGap(gap string) Printer {
	return &loggingPrinter{
		log: p.log,
		gap: p.gap + gap,
	}
}

func (p *loggingPrinter) Write(data []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.log == nil {
		return 0, nil
	}

	s := strings.Split(p.pending+string(data), "\n")
	if !strings.HasSuffix(string(data), "\n") {
		p.pending = s[len(s)-1]
	} else {
		p.pending = ""
	}
	lines := s[:len(s)-1]

	for _, l := range lines {
		p.log.Info(l)
	}

	return len(data), nil
}

func (p *loggingPrinter) Printf(msg string, args ...interface{}) (int, error) {
	return p.Write([]byte(fmt.Sprintf(msg, args...)))
}

func (p *loggingPrinter) Flush() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.pending != "" {
		p.log.Info(p.pending)
		p.pending = ""
	}
	return nil
}
