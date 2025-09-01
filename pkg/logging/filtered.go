package logging

import (
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/go-chi/chi/middleware"
)

// Logger defines the interface for logging operations
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// URLFilter defines the interface for URL filtering
type URLFilter interface {
	ShouldFilter(url string) bool
}

// RegexURLFilter implements URLFilter using regex patterns
type RegexURLFilter struct {
	pattern *regexp.Regexp
}

// ShouldFilter checks if the URL should be filtered
func (f *RegexURLFilter) ShouldFilter(url string) bool {
	return f.pattern != nil && f.pattern.MatchString(url)
}

// LoggingOption is a functional option for logging configuration
type LoggingOption func(*LoggingConfig)

// LoggingConfig holds configuration for request logging
type LoggingConfig struct {
	Logger    Logger
	Formatter middleware.LogFormatter
	URLFilter URLFilter
	NoColor   bool
	Output    io.Writer
}

// DefaultLoggingConfig provides sensible defaults
func DefaultLoggingConfig() *LoggingConfig {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	return &LoggingConfig{
		Logger: logger,
		Formatter: &middleware.DefaultLogFormatter{
			Logger:  logger,
			NoColor: false,
		},
		URLFilter: nil, // No filtering by default
		NoColor:   false,
		Output:    os.Stdout,
	}
}

// WithLogger sets a custom logger
func WithLogger(logger Logger) LoggingOption {
	return func(config *LoggingConfig) {
		config.Logger = logger
	}
}

// WithFormatter sets a custom log formatter
func WithFormatter(formatter middleware.LogFormatter) LoggingOption {
	return func(config *LoggingConfig) {
		config.Formatter = formatter
	}
}

// WithURLFilter sets a custom URL filter
func WithURLFilter(filter URLFilter) LoggingOption {
	return func(config *LoggingConfig) {
		config.URLFilter = filter
	}
}

// WithRegexFilter sets a regex-based URL filter
func WithRegexFilter(pattern *regexp.Regexp) LoggingOption {
	return func(config *LoggingConfig) {
		config.URLFilter = &RegexURLFilter{pattern: pattern}
	}
}

// WithNoColor disables color output
func WithNoColor(noColor bool) LoggingOption {
	return func(config *LoggingConfig) {
		config.NoColor = noColor
	}
}

// WithOutput sets the output writer
func WithOutput(output io.Writer) LoggingOption {
	return func(config *LoggingConfig) {
		config.Output = output
	}
}

// NewLoggingConfig creates a new logging config with options
func NewLoggingConfig(options ...LoggingOption) *LoggingConfig {
	config := DefaultLoggingConfig()
	for _, option := range options {
		option(config)
	}
	return config
}

// RequestLogger handles HTTP request logging with configuration
type RequestLogger struct {
	config *LoggingConfig
}

// NewRequestLogger creates a new request logger with options
func NewRequestLogger(options ...LoggingOption) *RequestLogger {
	config := NewLoggingConfig(options...)
	return &RequestLogger{config: config}
}

// Middleware returns the logging middleware function
func (rl *RequestLogger) Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Check if URL should be filtered
			if rl.config.URLFilter != nil && rl.config.URLFilter.ShouldFilter(r.URL.String()) {
				next.ServeHTTP(w, r)
				return
			}

			entry := rl.config.Formatter.NewLogEntry(r)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				entry.Write(ww.Status(), ww.BytesWritten(), ww.Header(), time.Since(t1), nil)
			}()

			next.ServeHTTP(ww, middleware.WithLogEntry(r, entry))
		}

		return http.HandlerFunc(fn)
	}
}

// Legacy functions for backward compatibility
func NewFilteredRequestLogger(filterOut *regexp.Regexp) func(next http.Handler) http.Handler {
	logger := NewRequestLogger(WithRegexFilter(filterOut))
	return logger.Middleware()
}

// FilteredRequestLogger is a copy of the middleware.RequestLogger function
// - But with a reg-ex to filter & exclude URLs from logging
func FilteredRequestLogger(f middleware.LogFormatter, urlRegEx *regexp.Regexp) func(next http.Handler) http.Handler {
	logger := NewRequestLogger(
		WithFormatter(f),
		WithRegexFilter(urlRegEx),
	)
	return logger.Middleware()
}
