package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/dnjp/zync/proto/zync/v1"
	"github.com/spf13/cobra"
)

func (c *client) removeFilesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm PATTERN",
		Short: "Remove files matching the given pattern stored in ipfs - not on the host",
		Args:  validRegexArg,
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.connect(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to connect to daemon: %+v\n", err)
				os.Exit(1)
			}

			cwd, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to read file: %+v\n", err)
				os.Exit(1)
			}

			pattern := args[0]
			if err := c.rm(cwd, pattern); err != nil {
				fmt.Fprintf(os.Stderr, "error removing files: %+v\n", err)
				os.Exit(1)
			}
		},
	}
}

func (c *client) rm(cwd, pattern string) error {

	fc, err := c.cc.DeleteFiles(context.TODO(), &zync.RegexRequest{
		Pattern:          pattern,
		CurrentDirectory: cwd,
	})
	if err != nil {
		return err
	}

	for {
		file, err := fc.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fmt.Fprintf(os.Stdout, "file: %+v\n", file)
	}

	return nil
}
