package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dnjp/zync/proto/zync/v1"
	"github.com/spf13/cobra"
)

func (c *client) backupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "backup",
		Short: "Initiates a full backup, returning the CID used for restores",
		Run: func(cmd *cobra.Command, args []string) {
			if err := c.connect(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to connect to daemon: %+v\n", err)
				os.Exit(1)
			}
			if err := c.backup(); err != nil {
				fmt.Fprintf(os.Stderr, "error during backup: %+v\n", err)
				os.Exit(1)
			}
		},
	}
}

func (c *client) backup() error {
	status, err := c.cc.Backup(context.TODO(), &zync.BackupRequest{})
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "status: %+v\n", status)

	return nil
}

