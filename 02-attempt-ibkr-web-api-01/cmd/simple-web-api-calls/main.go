package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"mih.ai/trading101/02-attempt-ibkr-web-api-01/clients"
)

func main() {
	err := run(context.Background(), os.Args, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(parentctx context.Context, args []string, stdin io.Reader, stdout io.Writer) error {
	_ = stdin
	logger, err := setupLogger(stdout)
	if err != nil {
		return fmt.Errorf("failed to set up the logger: %w", err)
	}
	logger.Info("Starting Simple API calls", slog.String("args", strings.Join(args, " ")))

	ctx, cancel := signal.NotifyContext(parentctx, os.Interrupt)
	defer cancel()

	httpClient := setupHTTPClientSkipVerify()

	cpgwClient := clients.NewCPGWClient(logger, httpClient, "https://localhost:5000/v1/api")

	// cpgwClient.Start_tickle(ctx)

	err = tryAPIs(cpgwClient)
	if err != nil {
		return fmt.Errorf("failed during tryAPIs(): %w", err)
	}

	<-ctx.Done()
	logger.Warn("exit run()", slog.String("error", ctx.Err().Error()), slog.String("cause", context.Cause(ctx).Error()))

	return nil
}

func setupLogger(stdout io.Writer) (*slog.Logger, error) {
	logger := slog.New(slog.NewJSONHandler(stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
	}))
	return logger, nil
}

func setupHTTPClientSkipVerify() *http.Client {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	return client
}

func tryAPIs(cpgwClient *clients.CPGWClient) error {
	// if err := cpgwClient.Do_iserver_auth_ssodh_init(); err != nil {
	// 	return err
	// }

	// if err := cpgwClient.Do_portfolio_accounts(); err != nil {
	// 	return err
	// }

	// if err := cpgwClient.Do_portfolio2_positions(); err != nil {
	// 	return err
	// }

	if err := cpgwClient.Do_iserver_account_orders(); err != nil {
		return err
	}

	if err := cpgwClient.Do_portfolio_positions_by_conid(848231171); err != nil {
		return err
	}

	return nil
}
