package network

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
	"time"
)

type Conn interface {
	io.ReadWriteCloser
}

type Stats struct {
	Sent     int64
	Recv     int64
	Duration time.Duration
}

type Bridge struct {
	src  Conn
	dest Conn
}

// NewBridge creates a new Bridge instance that bridges the given source and destination io.ReadWriteClosers.
// It returns a pointer to the created Bridge.
func NewBridge(src, dest Conn) *Bridge {
	return &Bridge{
		src:  src,
		dest: dest,
	}
}

// Run executes the bridge operation, copying data bidirectionally between the source and destination.
// It returns the statistics of the operation, including the number of bytes sent and received, and the duration of the operation.
// If any errors occur during the operation, it returns joined errors.
func (b *Bridge) Run(ctx context.Context) (Stats, error) {
	var sent, recv int64

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	startTime := time.Now()

	const ExpectedErrors = 3
	errCh := make(chan error, ExpectedErrors)

	startCopy(b.src, b.dest, &sent, errCh)
	startCopy(b.dest, b.src, &recv, errCh)

	go func() {
		<-ctx.Done()
		errCh <- b.Close()
	}()

	errs := make([]error, 0, ExpectedErrors)

	for i := 0; i < 3; i++ {
		if err := <-errCh; err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, syscall.ECONNRESET) {
			errs = append(errs, err)
		}

		cancel()
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

func startCopy(src io.Reader, dest io.Writer, sent *int64, out chan<- error) {
	go func() {
		var err error
		*sent, err = io.Copy(dest, src)
		out <- err
	}()
}

// Close closes the Bridge by closing the source and destination connections.
// It returns an error if there are any errors encountered while closing the connections.
// The function uses goroutines to close the connections concurrently and collects any errors that occur.
// If there are any errors, it returns a formatted error message with the joined errors.
// If there are no errors, it returns nil.
func (b *Bridge) Close() error {
	const ExpectedErrors = 2
	errsCh := make(chan error, ExpectedErrors)

	go func() { errsCh <- b.src.Close() }()
	go func() { errsCh <- b.dest.Close() }()

	errs := make([]error, 0, ExpectedErrors)

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
