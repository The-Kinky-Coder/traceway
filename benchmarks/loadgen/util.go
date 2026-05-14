package main

import "io"

func readAndDiscard(r io.Reader) (int64, error) {
	return io.Copy(io.Discard, r)
}
