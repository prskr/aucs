package testx

import (
	"context"
	"time"
)

type deadlined interface {
	Deadline() (deadline time.Time, ok bool)
	Cleanup(func())
}

func Context(d deadlined) context.Context {
	deadline, set := d.Deadline()
	if !set {
		ctx, cancel := context.WithCancel(context.Background())
		d.Cleanup(cancel)
		return ctx
	}

	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	d.Cleanup(cancel)

	return ctx
}
