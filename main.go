package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atye/chessfs/internal/domain"
	chessfs "github.com/atye/chessfs/internal/fs"
	"github.com/atye/chessfs/internal/lichess"
	golichess "github.com/atye/golichess/oapi-codegen"
	"github.com/hanwen/go-fuse/v2/fs"
)

func main() {
	debug := flag.Bool("debug", false, "print debug data")
	flag.Parse()

	logger := slog.Default()

	if len(flag.Args()) < 1 {
		logger.Error("usage:\n chessfs MOUNTPOINT")
		os.Exit(1)
	}

	lichessAccessToken := os.Getenv(domain.LichessTokenEnv)
	if lichessAccessToken == "" {
		logger.Error(fmt.Sprintf("env %s is empty", domain.LichessTokenEnv))
		os.Exit(1)
	}

	var chess domain.Chess

	httpClient := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &lichess.RateLimitRoundTripper{
			Proxied: http.DefaultTransport,
			Log:     logger,
		},
	}

	clientOps := []golichess.ClientOption{
		golichess.WithHTTPClient(httpClient),
		golichess.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", lichessAccessToken))
			if *debug {
				b, err := httputil.DumpRequest(req, true)
				if err != nil {
					return err
				}
				logger.Info("request", "dump", string(b))
			}
			return nil
		})}

	client, err := golichess.NewClient(domain.Lichess, clientOps...)
	if err != nil {
		logger.Error("creating lichess client", "error", err)
		os.Exit(1)
	}

	responseClient, err := golichess.NewClientWithResponses(domain.Lichess, clientOps...)
	if err != nil {
		logger.Error("creating lichess client", "error", err)
		os.Exit(1)
	}

	chess = &lichess.OAPICODEGEN{Client: client, ResponseClient: responseClient}

	opts := &fs.Options{}
	opts.Debug = *debug
	svr, err := fs.Mount(flag.Arg(0), chessfs.NewRoot(chess, logger), opts)
	if err != nil {
		logger.Error("mounting", "error", err)
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	go func() {
		<-sigCh
		if err := svr.Unmount(); err != nil {
			logger.Error("unmounting", "error", err)
		}
	}()

	svr.Wait()
}
