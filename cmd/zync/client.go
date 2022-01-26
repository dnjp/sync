package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/dnjp/zync/proto/zync/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const defaultRemote = "localhost:8081"

type client struct {
	remote string
	cc     zync.ZyncClient
}

func newClient() *client {
	return &client{}
}

func (c *client) connect() error {
	conn, err := grpc.DialContext(
		context.Background(),
		c.remote,
		grpc.WithInsecure(),
	)
	if err != nil {
		return err
	}
	c.cc = zync.NewZyncClient(conn)
	return nil
}

func (c *client) rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "zync COMMAND",
		DisableFlagsInUseLine: false,
		Args:                  cobra.MinimumNArgs(1),
		Short:                 "command line client for the zyncd daemon",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	c.initCommands(cmd)
	c.initFlags(cmd)
	return cmd
}

func (c *client) initCommands(cmd *cobra.Command) {
	cmd.AddCommand(c.addFilesCmd())
	cmd.AddCommand(c.listFilesCmd())
	cmd.AddCommand(c.removeFilesCmd())
	cmd.AddCommand(c.backupCmd())
	cmd.AddCommand(c.restoreCmd())
}

func (c *client) initFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&c.remote, "url", defaultRemote, "the url to the daemon")
}

func validRegexArg(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing file pattern")
	}
	_, err := regexp.Compile(args[0])
	return err
}
