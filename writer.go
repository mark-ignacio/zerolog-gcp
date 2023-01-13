package zlg

import (
	"context"
	"encoding/json"
	"sync"

	"cloud.google.com/go/logging"
	"github.com/rs/zerolog"
	"google.golang.org/api/option"
)

type cloudLoggingWriter struct {
	ctx                context.Context
	firstBlockingWrite sync.Once
	logger             *logging.Logger
	severityMap        map[zerolog.Level]logging.Severity

	zerolog.LevelWriter
}

// DefaultSeverityMap contains the default zerolog.Level -> logging.Severity mappings.
var DefaultSeverityMap = map[zerolog.Level]logging.Severity{
	zerolog.DebugLevel: logging.Debug,
	zerolog.InfoLevel:  logging.Info,
	zerolog.WarnLevel:  logging.Warning,
	zerolog.ErrorLevel: logging.Error,
	zerolog.PanicLevel: logging.Critical,
	zerolog.FatalLevel: logging.Critical,
}

// secretly, we keep tabs of all loggers
var loggersWeMade = make([]*logging.Logger, 0, 1)

func (c *cloudLoggingWriter) Write(p []byte) (int, error) {
	// writing to stackdriver without levels? o-okay...
	entry := logging.Entry{Payload: json.RawMessage(p)}
	c.logger.Log(entry)
	var err error
	c.firstBlockingWrite.Do(func() {
		err = c.logger.Flush()
	})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *cloudLoggingWriter) WriteLevel(level zerolog.Level, payload []byte) (int, error) {
	entry := logging.Entry{
		Severity: c.severityMap[level],
		Payload:  json.RawMessage(payload),
	}
	c.logger.Log(entry)
	var err error
	c.firstBlockingWrite.Do(func() {
		err = c.logger.Flush()
	})
	if err != nil {
		return 0, err
	}
	if level == zerolog.FatalLevel {
		// ensure that any pending logs are written before exit
		err := c.logger.Flush()
		if err != nil {
			return 0, err
		}
	}
	return len(payload), nil
}

// CloudLoggingOptions specifies some optional configuration.
type CloudLoggingOptions struct {
	// Specify this to override DefaultSeverityMap.
	SeverityMap map[zerolog.Level]logging.Severity

	// Used during *logging.Client construction.
	ClientOptions []option.ClientOption

	// Used during *logging.Client construction.
	ClientOnError func(error)

	// Specify this to override the default of constructing a *logging.Logger on the caller's behalf.
	Logger *logging.Logger

	// Used during GCP Logger construction.
	LoggerOptions []logging.LoggerOption
}

// NewCloudLoggingWriter creates a LevelWriter that logs only to GCP Cloud Logging using non-blocking calls.
func NewCloudLoggingWriter(ctx context.Context, projectID, logID string, opts CloudLoggingOptions) (writer zerolog.LevelWriter, err error) {
	logger := opts.Logger
	if opts.Logger == nil {
		var client *logging.Client
		client, err = logging.NewClient(ctx, projectID, opts.ClientOptions...)
		if err != nil {
			return
		}
		if opts.ClientOnError != nil {
			client.OnError = opts.ClientOnError
		}
		logger = client.Logger(logID, opts.LoggerOptions...)
		loggersWeMade = append(loggersWeMade, logger)
	}
	severityMap := opts.SeverityMap
	if severityMap == nil {
		severityMap = DefaultSeverityMap
	}
	writer = &cloudLoggingWriter{
		ctx:         ctx,
		logger:      logger,
		severityMap: severityMap,
	}
	return
}

// Flush blocks while flushing all loggers this module created.
func Flush() []error {
	var errs []error
	for _, logger := range loggersWeMade {
		if logger != nil {
			if err := logger.Flush(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
}
