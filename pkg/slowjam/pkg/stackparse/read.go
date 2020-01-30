package stackparse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/maruel/panicparse/stack"
)

type StackSample struct {
	Time    time.Time
	Context *stack.Context
}

// Read parses a stack log input
func Read(r io.Reader) ([]*StackSample, error) {
	inStack := false
	t := time.Time{}
	sd := bytes.NewBuffer([]byte{})
	samples := []*StackSample{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if !inStack {
			line := scanner.Text()
			// 	fmt.Printf("ts: %v\n", line)
			s, err := strconv.ParseInt(line, 10, 64)
			if err != nil {
				return samples, err
			}
			t = time.Unix(0, s)
			inStack = true
			continue
		}
		if strings.HasPrefix(scanner.Text(), "-") {
			// 	fmt.Printf("end stack marker\n")
			inStack = false
			ctx, err := stack.ParseDump(sd, os.Stdout, false)
			if err != nil {
				fmt.Printf("parse err: %v", err)
				return samples, err
			}
			// 	fmt.Printf("\nCONTEXT: %+v\n", ctx)
			samples = append(samples, &StackSample{Time: t, Context: ctx})
			continue
		}
		sd.Write(scanner.Bytes())
		sd.Write([]byte{'\n'})
	}

	if err := scanner.Err(); err != nil {
		return samples, err
	}
	return samples, nil
}
