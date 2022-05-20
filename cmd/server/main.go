// Copyright (c) 2022 Yuriy Iovkov

package server

import (
	"context"
	"errors"
	"os"

	"github.com/rurick/powprotected/internal/app/server"
	"github.com/rurick/powprotected/pkg/dotenv"
	"github.com/rurick/powprotected/pkg/shutdown"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	envTCPAddr     = "POW_APP_TCP_ADDRESS"
	defaultTCPAddr = ":8888"
)

func main() {
	dotenv.Overload()
	logger := logrus.New()
	addr, ok := os.LookupEnv(envTCPAddr)
	if !ok {
		logger.Info("env not found:", envTCPAddr)
		addr = defaultTCPAddr
	}

	if err := run(logger, addr); err != nil {
		logger.Fatal(err)
	}
}

func run(logger *logrus.Logger, addr string) error {
	eg, ctx := errgroup.WithContext(context.Background())

	srv := server.New(addr, logger)
	eg.Go(func() error {
		return srv.Start(ctx)
	})

	// обработка сигналов ОС
	sigHandler := shutdown.TermSignalTrap()
	eg.Go(func() error {
		err := sigHandler.Wait(ctx)
		srv.Stop()
		return err
	})

	if err := eg.Wait(); err != nil &&
		!errors.Is(err, shutdown.ErrTermSig) &&
		!errors.Is(err, context.Canceled) {
		return err
	}
	logger.Info("graceful shutdown successfully finished")
	return nil
}
