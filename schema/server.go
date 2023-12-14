package schema

import (
	"io"
)

const BROADCAST_MESSAGE = "*"

type Status string

const STATUS_ENQUEUED Status = "enqueued"
const STATUS_WAITING Status = "waiting"
const STATUS_RUNNING Status = "running"
const STATUS_SUCCESS Status = "success"
const STATUS_FAILED Status = "failed"
const STATUS_TIMEOUT Status = "timeout"

type Error string

func (err Error) Error() string {
	return string(err)
}

const ERROR_UNAVAILABLE = Error("not available")

const EVENT_STARTUP_COMPLETE = "startup complete"

const MESSAGE_SOURCE_SERVER = "*server"
const MESSAGE_SOURCE_API = "*api"

func IsMessageFromPlugin(source string) bool {
	switch source {
	case MESSAGE_SOURCE_SERVER, MESSAGE_SOURCE_API:
		return false

	default:
		return true
	}
}

type Message struct {
	Target  string            `json:"target"`
	Options map[string]string `json:"options"`
	Data    []byte            `json:"data"`
}

func BroadcastMessage(options map[string]string, data []byte) Message {
	return Message{Target: BROADCAST_MESSAGE, Options: options, Data: data}
}

type FullMessage struct {
	Message
	Source string
}

type Trigger map[string]string

type PipelineStatus struct {
	Pipeline    Pipeline
	WorkerGroup string
	ActivityID  string

	Status Status
	Logs   LogReaderProvider
	Result PipelineResult
}

func (s *PipelineStatus) Running() bool {
	switch s.Status {
	case STATUS_WAITING, STATUS_RUNNING:
		return true

	default:
		return false
	}
}

func (s *PipelineStatus) Finished() bool {
	switch s.Status {
	case STATUS_SUCCESS, STATUS_FAILED, STATUS_TIMEOUT:
		return true

	default:
		return false
	}
}

type LogReader interface {
	io.ReadSeekCloser

	ReadAt(p []byte, offset int64) (n int, err error)
	Size() (int64, bool)
}

type LogReaderProvider interface {
	Available() bool
	Reader() (LogReader, error)
	io.Closer
}

type WorkerQueueResponse struct {
	Contract string   `json:"contract"`
	Activity string   `json:"activity"`
	Pipeline Pipeline `json:"pipeline"`
}

type WorkerAckRequest struct {
	Contract string `json:"contract"`
}

type WorkerLogsPositionResponse struct {
	Position int64 `json:"position"`
}

type PipelineResult struct {
	Success  bool   `json:"success"`
	ExitCode int    `json:"exitCode"`
	Error    string `json:"error"`
}
