package zlg

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/logging"
	"github.com/rs/zerolog"
)

type cloudLoggingWriter struct {
	ctx         context.Context
	wroteOnce   bool
	logger      *logging.Logger
	severityMap map[zerolog.Level]logging.Severity

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

func (c *cloudLoggingWriter) Write(p []byte) (int, error) {
	// writing to stackdriver without levels? o-okay...
	entry := logging.Entry{Payload: json.RawMessage(p)}
	if !c.wroteOnce {
		c.wroteOnce = true
		err := c.logger.LogSync(c.ctx, entry)
		if err != nil {
			return 0, err
		}
	} else {
		c.logger.Log(entry)
	}
	return len(p), nil
}

func (c *cloudLoggingWriter) WriteLevel(level zerolog.Level, payload []byte) (int, error) {
	entry := logging.Entry{
		Severity: c.severityMap[level],
		Payload:  json.RawMessage(payload),
	}
	if !c.wroteOnce {
		c.wroteOnce = true
		err := c.logger.LogSync(c.ctx, entry)
		if err != nil {
			return 0, err
		}
	} else {
		c.logger.Log(entry)
	}
	return len(payload), nil
}

// CloudLoggingOptions specifies some optional configuration.
type CloudLoggingOptions struct {
	// SeverityMap can be optionally specified to use instead of DefaultSeverityMap.
	SeverityMap map[zerolog.Level]logging.Severity

	// Logger can be optionally provided in lieu of constructing a logger on the caller's behalf.
	Logger *logging.Logger

	// LoggerOptions is optionally used to construct a Logger.
	LoggerOptions []logging.LoggerOption
}

// NewCloudLoggingWriter creates a LevelWriter that logs only to GCP Cloud Logging using non-blocking calls.
func NewCloudLoggingWriter(ctx context.Context, projectID, logID string, opts CloudLoggingOptions) (writer zerolog.LevelWriter, err error) {
	logger := opts.Logger
	if opts.Logger == nil {
		var client *logging.Client
		client, err = logging.NewClient(ctx, projectID)
		if err != nil {
			return
		}
		logger = client.Logger(logID, opts.LoggerOptions...)
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
