package main

import (
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"mih.ai/trading101/01-attempt-ibapi-go/clients"
)

const (
	IB_HOST = "127.0.0.1"
	IB_PORT = 7497
)

func main() {
	err := run(os.Args, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	_ = stdin
	logger, err := setupLogger(stdout)
	if err != nil {
		return fmt.Errorf("failed to set up the logger: %w", err)
	}
	logger.Info("Starting Simple API calls", slog.String("args", strings.Join(args, " ")))

	ibapiEClient := setupIBAPIEClient(stdout)
	defer ibapiEClient.Disconnect()

	ibapiEClient.Connect(IB_HOST, IB_PORT, rand.Int64N(999999))

	return nil
}

func setupLogger(stdout io.Writer) (*slog.Logger, error) {
	return slog.New(slog.NewJSONHandler(stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Align time format with zerolog
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				return slog.String(slog.TimeKey, t.Format(time.RFC3339))
			}
			return a
		},
	})), nil
}

func setupIBAPIEClient(logwriter io.Writer) *clients.IBAPIEClient {
	logger := zerolog.New(logwriter).Level(zerolog.InfoLevel).With().Timestamp().Caller().Logger()
	// Align with slog
	zerolog.MessageFieldName = "msg"
	zerolog.CallerFieldName = "source"
	// zerolog.TimeFieldFormat = time.RFC3339Nano

	logger.Info().Msg("Setup IBAPI EClient")
	return clients.NewIBAPIEClient(logger)
}
