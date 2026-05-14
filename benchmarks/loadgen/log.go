package main

import (
	"io"
	"os"
	"sync"
)

var stderrMu sync.Mutex

// stderrPrefix returns a single shared writer for human-readable progress logs.
// Wrapping os.Stderr through a tiny adapter so concurrent goroutines never
// interleave half a line — important because the ramp loop and ingest worker
// pool both want to print status updates while running.
func stderrPrefix() io.Writer {
	return &lockedStderr{}
}

type lockedStderr struct{}

func (l *lockedStderr) Write(p []byte) (int, error) {
	stderrMu.Lock()
	defer stderrMu.Unlock()
	return os.Stderr.Write(p)
}
