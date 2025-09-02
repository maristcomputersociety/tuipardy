package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s <board.csv>\n", os.Args[0])
	}
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}
	csvPath := flag.Arg(0)

	board, err := LoadBoard(csvPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading board: %v\n", err)
		os.Exit(1)
	}

	g := NewGame(board)
	if err := g.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}
