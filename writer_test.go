package zlg

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"cloud.google.com/go/logging"
	"github.com/rs/zerolog"
)

// makeRedirectLogger creates a logger that redirects JSON output to buf,
// bypassing real GCP calls. Returns nil if no GCP credentials are available.
func makeRedirectLogger(buf *bytes.Buffer) *logging.Logger {
	ctx := context.Background()
	client, err := logging.NewClient(ctx, "test-project")
	if err != nil {
		return nil
	}
	return client.Logger("test-log", logging.RedirectAsJSON(buf))
}

func newWriterFromLogger(logger *logging.Logger) (zerolog.LevelWriter, error) {
	return NewCloudLoggingWriter(context.Background(), "test-project", "test-log", &CloudLoggingOptions{
		Logger: logger,
	})
}

// --- SeverityMap tests ---

func TestNewCloudLoggingWriter_SeverityMap(t *testing.T) {
	logger := makeRedirectLogger(&bytes.Buffer{})
	if logger == nil {
		t.Skip("skipping: no GCP credentials")
	}

	t.Run("default", func(t *testing.T) {
		w, err := newWriterFromLogger(logger)
		if err != nil {
			t.Fatal(err)
		}

		c := w.(*cloudLoggingWriter)
		for level, want := range DefaultSeverityMap {
			if got := c.severityMap[level]; got != want {
				t.Errorf("severityMap[%s] = %q, want %q", level, got, want)
			}
		}
	})

	t.Run("nil uses default", func(t *testing.T) {
		w, err := NewCloudLoggingWriter(context.Background(), "test-project", "test-log", &CloudLoggingOptions{
			Logger:      logger,
			SeverityMap: nil,
		})
		if err != nil {
			t.Fatal(err)
		}

		c := w.(*cloudLoggingWriter)
		if c.severityMap[zerolog.ErrorLevel] != logging.Error {
			t.Errorf("expected default ErrorLevel -> Error, got %v", c.severityMap[zerolog.ErrorLevel])
		}
		if c.severityMap[zerolog.DebugLevel] != logging.Debug {
			t.Errorf("expected default DebugLevel -> Debug, got %v", c.severityMap[zerolog.DebugLevel])
		}
	})

	t.Run("WriteLevel applies severity", func(t *testing.T) {
		var buf bytes.Buffer
		logger := makeRedirectLogger(&buf)
		if logger == nil {
			t.Skip("skipping: no GCP credentials")
		}

		w, err := NewCloudLoggingWriter(context.Background(), "test-project", "test-log", &CloudLoggingOptions{
			Logger:      logger,
			SeverityMap: map[zerolog.Level]logging.Severity{zerolog.InfoLevel: logging.Alert},
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = w.WriteLevel(zerolog.InfoLevel, []byte(`{"message":"info-test"}`))
		if err != nil {
			t.Fatalf("WriteLevel: %v", err)
		}

		if buf.Len() == 0 {
			t.Fatal("expected output in redirect buffer")
		}

		for line := range bytes.SplitSeq(buf.Bytes(), []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var entry struct {
				Severity string `json:"severity"`
			}
			if err := json.Unmarshal(line, &entry); err != nil {
				continue
			}
			if entry.Severity == "ALERT" {
				return
			}
		}
		t.Fatalf("expected ALERT severity in output, got:\n%s", buf.String())
	})
}
