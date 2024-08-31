package network

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

type Stats struct {
	Sent     int64
	Recv     int64
	Duration time.Duration
}

type Bridge struct {
	src  io.ReadWriteCloser
	dest io.ReadWriteCloser
}

func NewBridge(src, dest io.ReadWriteCloser) *Bridge {
	return &Bridge{
		src:  src,
		dest: dest,
	}
}

func (b *Bridge) Run(ctx context.Context) (Stats, error) {
	var sent, recv int64

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	startTime := time.Now()
	errCh := make(chan error, 3)

	go func() {
		defer cancel()
		var err error

		sent, err = io.Copy(b.src, b.dest)
		errCh <- err
	}()

	go func() {
		defer cancel()
		var err error

		recv, err = io.Copy(b.dest, b.src)
		errCh <- err
	}()

	go func() {
		<-ctx.Done()
		errCh <- b.src.Close()
	}()

	errs := make([]error, 0, 3)
	for i := 0; i < 3; i++ {
		if err := <-errCh; err != nil {
			errs = append(errs, err)
		}
	}

	var err error
	if len(errs) > 0 {
		err = fmt.Errorf("error to run bridge: %w", errors.Join(errs...))
	}

	return Stats{
		Sent:     sent,
		Recv:     recv,
		Duration: time.Since(startTime),
	}, err
}

func (b *Bridge) Close() error {
	errsCh := make(chan error, 2)

	go func() { errsCh <- b.src.Close() }()
	go func() { errsCh <- b.dest.Close() }()

	errs := make([]error, 0, 2)
	for i := 0; i < 2; i++ {
		if err := <-errsCh; err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("error to close connections: %w", errors.Join(errs...))
	}

	return nil
}
