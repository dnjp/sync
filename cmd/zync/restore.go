package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/dnjp/zync/proto/zync/v1"
	"github.com/spf13/cobra"
)

func (c *client) restoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restore CID",
		Args:  cobra.MinimumNArgs(1),
		Short: "Initiates a full restore from the database held at the given CID",
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.connect(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to connect to daemon: %+v\n", err)
				os.Exit(1)
			}

			if err := c.restore(args[0]); err != nil {
				fmt.Fprintf(os.Stderr, "error during restore: %+v\n", err)
				os.Exit(1)
			}
		},
	}
}

func (c *client) restore(cid string) error {
	rc, err := c.cc.Restore(context.TODO(), &zync.RestoreRequest{
		Cid: cid,
	})
	if err != nil {
		return err
	}

	for {
		update, err := rc.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fmt.Fprintf(os.Stdout, "update %+v\n", update)
	}

	return nil
}

