package reader

import (
	"bufio"
	"context"
	"io"
	"os"
)

type StdinReader struct {
	scanner *bufio.Scanner
}

func NewStdinReader() *StdinReader {
	return &StdinReader{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

func (r *StdinReader) ReadLines(ctx context.Context, output chan<- string) error {
	defer close(output)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if !r.scanner.Scan() {
				if err := r.scanner.Err(); err != nil && err != io.EOF {
					return err
				}
				return nil
			}

			line := r.scanner.Text()
			if line != "" {
				select {
				case output <- line:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
}
