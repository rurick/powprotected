// Copyright (c) 2022 Yuriy Iovkov

package client

import (
	"context"
	"errors"

	"github.com/rurick/powprotected/pkg/dotenv"
	"github.com/rurick/powprotected/pkg/shutdown"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {
	dotenv.Overload()
	log := logrus.New()
	if err := run(log); err != nil {
		log.Fatal(err)
	}
}

func run(log *logrus.Logger) error {
	eg, ctx := errgroup.WithContext(context.Background())

	// обработка сигналов ОС
	sigHandler := shutdown.TermSignalTrap()
	eg.Go(func() error {
		return sigHandler.Wait(ctx)
	})

	if err := eg.Wait(); err != nil &&
		!errors.Is(err, shutdown.ErrTermSig) &&
		!errors.Is(err, context.Canceled) {
		return err
	}
	log.Info("graceful shutdown successfully finished")
	return nil
}
