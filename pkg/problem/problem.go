package problem

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
)

// Logger defines the interface for logging operations
type Logger interface {
	Printf(format string, v ...interface{})
}

// DefaultLogger implements Logger using the standard log package
type DefaultLogger struct{}

// Printf logs using the standard log package
func (l *DefaultLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// ProblemOption is a functional option for problem configuration
type ProblemOption func(*ProblemConfig)

// ProblemConfig holds configuration for problem responses
type ProblemConfig struct {
	Logger    Logger
	LogPrefix string
	LogErrors bool
}

// DefaultProblemConfig provides sensible defaults
func DefaultProblemConfig() *ProblemConfig {
	return &ProblemConfig{
		Logger:    &DefaultLogger{},
		LogPrefix: "### 💥 API",
		LogErrors: true,
	}
}

// WithLogger sets a custom logger
func WithLogger(logger Logger) ProblemOption {
	return func(config *ProblemConfig) {
		config.Logger = logger
	}
}

// WithLogPrefix sets the log prefix
func WithLogPrefix(prefix string) ProblemOption {
	return func(config *ProblemConfig) {
		config.LogPrefix = prefix
	}
}

// WithLogErrors enables/disables error logging
func WithLogErrors(logErrors bool) ProblemOption {
	return func(config *ProblemConfig) {
		config.LogErrors = logErrors
	}
}

// NewProblemConfig creates a new problem config with options
func NewProblemConfig(options ...ProblemOption) *ProblemConfig {
	config := DefaultProblemConfig()
	for _, option := range options {
		option(config)
	}
	return config
}

// ProblemManager handles problem response creation and configuration
type ProblemManager struct {
	config *ProblemConfig
}

// NewProblemManager creates a new problem manager with options
func NewProblemManager(options ...ProblemOption) *ProblemManager {
	config := NewProblemConfig(options...)
	return &ProblemManager{config: config}
}

type Problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// New creates a new problem with the manager's configuration
func (pm *ProblemManager) New(typeStr string, title string, status int, detail, instance string) *Problem {
	return &Problem{typeStr, title, status, detail, instance}
}

// Send sends the problem response with logging
func (pm *ProblemManager) Send(p *Problem, resp http.ResponseWriter) {
	if pm.config.LogErrors {
		pm.config.Logger.Printf("%s %s", pm.config.LogPrefix, p.Error())
	}
	resp.Header().Set("Content-Type", "application/problem+json")
	resp.WriteHeader(p.Status)
	_ = json.NewEncoder(resp).Encode(p)
}

// Wrap wraps an error into a problem response
func (pm *ProblemManager) Wrap(status int, typeStr string, instance string, err error) *Problem {
	var p *Problem
	if err != nil {
		p = pm.New(typeStr, MyCaller(), status, err.Error(), instance)
	} else {
		p = pm.New(typeStr, MyCaller(), status, "Other error occurred", instance)
	}

	return p
}

// Legacy functions for backward compatibility
func New(typeStr string, title string, status int, detail, instance string) *Problem {
	manager := NewProblemManager()
	return manager.New(typeStr, title, status, detail, instance)
}

func (p *Problem) Send(resp http.ResponseWriter) {
	manager := NewProblemManager()
	manager.Send(p, resp)
}

func Wrap(status int, typeStr string, instance string, err error) *Problem {
	manager := NewProblemManager()
	return manager.Wrap(status, typeStr, instance, err)
}

func (p Problem) Error() string {
	return fmt.Sprintf("Problem: Type: '%s', Title: '%s', Status: '%d', Detail: '%s', Instance: '%s'",
		p.Type, p.Title, p.Status, p.Detail, p.Instance)
}

func getFrame(skipFrames int) runtime.Frame {
	targetFrameIndex := skipFrames + 2

	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}

	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])

		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()

			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}

func MyCaller() string {
	return getFrame(2).Function
}
