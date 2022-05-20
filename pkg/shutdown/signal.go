package shutdown

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
)

var (
	// ErrTermSig возвращается, когда от ОС получен один из следующих сигналов: SIGINT, SIGTERM
	ErrTermSig = errors.New("termination signal caught")
)

// SignalTrap используется для перехвата сигнала от ОС и инициации
// процесса завершения работы, используя инструменты оркестрации (Например: golang.org/x/sync/errgroup).
type SignalTrap chan os.Signal

// TermSignalTrap возвращает SignalTrap, который слушает сигналы от ОС.
func TermSignalTrap() SignalTrap {
	trap := SignalTrap(make(chan os.Signal, 1))

	signal.Notify(trap, syscall.SIGINT, os.Interrupt, syscall.SIGTERM)

	return trap
}

// Wait ожидает сигнал от ОС и возвращает ErrTermSig, если был получен сигнал завершения работы.
// Блокирует поток выполнения.
func (t SignalTrap) Wait(ctx context.Context) error {
	select {
	case <-t:
		return ErrTermSig
	case <-ctx.Done():
		return ctx.Err()
	}
}
