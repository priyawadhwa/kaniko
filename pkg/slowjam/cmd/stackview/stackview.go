package main

import (
	"fmt"
	"slowjam/pkg/stackparse"
	"os"
	"strings"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(fmt.Sprintf("open: %v", err))
	}
	defer f.Close()
	samples, err := stackparse.Read(f)
	if err != nil {
		panic(fmt.Sprintf("parse: %v", err))
	}

	tl := stackparse.CreateTimeline(samples, stackparse.SuggestedIgnore)

	fmt.Printf("%d samples over %s\n", tl.Samples, tl.End.Sub(tl.Start))
	for _, g := range tl.Goroutines {
		fmt.Printf("goroutine %d (%s)\n", g.ID, g.Signature.CreatedByString(true))
		for i, l := range g.Layers {
			for _, c := range l.Calls {
				if c.Samples > 1 {
					fmt.Printf(" %s %s execution time: %s (%d samples)\n", strings.Repeat(" ", i), c.Name, c.EndDelta-c.StartDelta, c.Samples)
				}
			}
		}
		fmt.Printf("\n")
	}
}
