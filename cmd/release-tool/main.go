package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/lwsanty/release-tool/cmd/release-tool/cmd"
)

func main() {
	parser := flags.NewNamedParser("release-tool", flags.Default)
	parser.AddCommand("collect", cmd.CollectCommandDescription, "", &cmd.CollectCommand{})
	parser.AddCommand("apply", cmd.ApplyCommandDescription, "", &cmd.ApplyCommand{})

	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
}
