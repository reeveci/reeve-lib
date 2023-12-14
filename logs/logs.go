package logs

import (
	"fmt"
	"io"
	"sync"
)

func New(target io.Writer, name string) LogWriter {
	return NewDecorated(target, name, NewDefaultDecorator("", ""))
}

func NewDecorated(target io.Writer, name string, decorator Decorator) LogWriter {
	w := &rootLogWriter{
		target:    target,
		scope:     name,
		decorator: decorator,
		newline:   '\n',
	}
	w.writerPrinter = writerPrinter{w}
	return w
}

type LogWriter interface {
	io.Writer
	io.StringWriter
	Printer
	Subsystem(name string) LogWriter
}

type Printer interface {
	Print(a ...interface{}) (int, error)
	Printf(format string, a ...interface{}) (int, error)
	Println(values ...interface{}) (int, error)
}

type ScopeWriter interface {
	GetScope(scope string, name string) string
	WriteScope(scope string, b []byte) (int, error)
}

type LineContext struct {
	// The scope of the current line
	Scope string

	// This line is interrupting a previous line
	Interrupting bool
	// The scope of the interrupted previous line
	InterruptedScope string

	// This line is going to be interrupted by the next line
	Interrupted bool
	// The scope of the interrupting next line
	InterruptingScope string
}

type rootLogWriter struct {
	target    io.Writer
	scope     string
	decorator Decorator

	context  LineContext
	position int
	newline  byte

	writerPrinter
	sync.Mutex
}

func (w *rootLogWriter) Write(b []byte) (n int, err error) {
	return w.WriteScope(w.scope, b)
}

func (w *rootLogWriter) WriteString(s string) (n int, err error) {
	return w.WriteScope(w.scope, []byte(s))
}

func (w *rootLogWriter) Subsystem(name string) LogWriter {
	return newSubsystemLogWriter(w, w.GetScope(w.scope, name))
}

func (w *rootLogWriter) GetScope(scope, name string) string {
	return w.decorator.Subsystem(scope, name)
}

const start = 0
const middle = 1
const end = 2

func (w *rootLogWriter) WriteScope(scope string, b []byte) (int, error) {
	w.Lock()
	defer w.Unlock()

	context := w.context
	position := w.position

	// are we interrupting another scope?
	if w.context.Scope != scope && position != start {
		context.Interrupted = true
		context.InterruptingScope = scope

		if position == middle {
			position = end

			suffix := w.decorator.LineSuffix(context)
			if suffix != "" {
				n, err := w.target.Write([]byte(suffix))
				if n > 0 {
					w.context = context
					w.position = position
				}
				if err != nil {
					return 0, err
				}
			}
		}
	}

	switch position {
	case start, middle:
		context = LineContext{
			Scope:            scope,
			Interrupting:     context.Interrupting,
			InterruptedScope: context.InterruptedScope,
		}

	case end:
		context = LineContext{
			Scope:            scope,
			Interrupting:     context.Interrupted,
			InterruptedScope: context.Scope,
		}
		position = start

		n, err := w.target.Write([]byte{w.newline})
		if n > 0 {
			w.context = context
			w.position = position
		}
		if err != nil {
			return 0, err
		}
	}

	buffer := make([]byte, 0, len(b))
	var written int

	flush := func() error {
		if len(buffer) == 0 {
			return nil
		}
		n, err := w.target.Write(buffer)
		if n > 0 {
			written += n
			w.context = context
			if n <= len(buffer) && buffer[n-1] == w.newline {
				w.position = start
			} else {
				w.position = middle
			}
		}
		if err != nil {
			return err
		}
		buffer = buffer[:0]
		return nil
	}

	for _, char := range b {
		if position == start {
			position = middle

			prefix := w.decorator.LinePrefix(context)
			if prefix != "" {
				if err := flush(); err != nil {
					return written, err
				}

				n, err := w.target.Write([]byte(prefix))
				if n > 0 {
					w.context = context
					w.position = position
				}
				if err != nil {
					return written, err
				}
			}
		}

		if char == w.newline {
			position = end

			suffix := w.decorator.LineSuffix(context)
			if suffix != "" {
				if err := flush(); err != nil {
					return written, err
				}
				n, err := w.target.Write([]byte(suffix))
				if n > 0 {
					w.context = context
					w.position = position
				}
				if err != nil {
					return written, err
				}
			}

			context = LineContext{Scope: scope}
			position = start
		}

		buffer = append(buffer, char)
	}

	err := flush()
	return written, err
}

func newSubsystemLogWriter(parent ScopeWriter, scope string) LogWriter {
	w := subsystemLogWriter{parent: parent, scope: scope}
	w.writerPrinter = writerPrinter{w}
	return w
}

type subsystemLogWriter struct {
	parent ScopeWriter
	scope  string

	writerPrinter
}

func (w subsystemLogWriter) Write(b []byte) (n int, err error) {
	return w.WriteScope(w.scope, b)
}

func (w subsystemLogWriter) WriteString(s string) (n int, err error) {
	return w.WriteScope(w.scope, []byte(s))
}

func (w subsystemLogWriter) Subsystem(name string) LogWriter {
	return newSubsystemLogWriter(w, w.GetScope(w.scope, name))
}

func (w subsystemLogWriter) GetScope(scope, name string) string {
	return w.parent.GetScope(scope, name)
}

func (w subsystemLogWriter) WriteScope(scope string, b []byte) (n int, err error) {
	return w.parent.WriteScope(scope, b)
}

type writerPrinter struct {
	target io.Writer
}

func (w writerPrinter) Print(a ...any) (int, error) {
	return fmt.Fprint(w.target, a...)
}

func (w writerPrinter) Printf(format string, a ...any) (int, error) {
	return fmt.Fprintf(w.target, format, a...)
}

func (w writerPrinter) Println(a ...any) (int, error) {
	return fmt.Fprintln(w.target, a...)
}

type Decorator interface {
	// Get the unique name for a subsystem for the specified scope
	Subsystem(scope, name string) string
	// Get the prefix for a line
	LinePrefix(context LineContext) string
	// Get the suffix for a line
	LineSuffix(context LineContext) string
}

func NewDefaultDecorator(marker, prefixTemplate string) Decorator {
	return defaultDecorator{marker: marker, prefixTemplate: prefixTemplate, interrupted: map[string]bool{}}
}

type defaultDecorator struct {
	marker, prefixTemplate string
	interrupted            map[string]bool
}

func (d defaultDecorator) Subsystem(scope, name string) string {
	if scope == "" {
		return name
	}
	return scope + ":" + name
}

func (d defaultDecorator) LinePrefix(context LineContext) string {
	if context.Scope == "" {
		return ""
	}

	prefix := "[" + context.Scope + d.marker + "]"
	if d.prefixTemplate != "" {
		prefix = fmt.Sprintf(d.prefixTemplate, prefix)
	}

	if d.interrupted[context.Scope] {
		delete(d.interrupted, context.Scope)
		prefix += "…"
	} else {
		prefix += " "
	}

	return prefix
}

func (d defaultDecorator) LineSuffix(context LineContext) string {
	if context.Scope == "" {
		return ""
	}

	if context.Interrupted {
		d.interrupted[context.Scope] = true
		return "…"
	} else {
		return ""
	}
}
