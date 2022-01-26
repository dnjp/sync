package main

import (
	"fmt"
	"log"
	"os"
	"syscall"

	zyncd "github.com/dnjp/zync/daemon"
	"github.com/sevlyar/go-daemon"
	"github.com/spf13/cobra"
)

func newDaemon() *daemon.Context {
	return &daemon.Context{
		PidFileName: "zyncd.pid",
		PidFilePerm: 0644,
		LogFileName: "zyncd.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{},
	}
}

func startDaemon(ctx *daemon.Context) (proc *os.Process, shouldExit bool, err error) {
	proc, err = ctx.Reborn()
	if err != nil {
		return
	}
	if proc != nil {
		shouldExit = true
		return
	}
	return
}

func stopDaemon(ctx *daemon.Context) error {
	return ctx.Release()
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "zyncd COMMAND",
		DisableFlagsInUseLine: false,
		Args:                  cobra.MinimumNArgs(1),
		Short:                 "command line client for the zyncd daemon",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	initCommands(cmd)
	return cmd
}

func initCommands(cmd *cobra.Command) {
	cmd.AddCommand(startCmd())
}

func startCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Launches the daemon",
		Run: func(cmd *cobra.Command, args []string) {

			server, err := zyncd.NewServer(8081)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%+v\n", err)
				os.Exit(1)
			}

			daemon.SetSigHandler(
				server.HandleTerminate,
				syscall.SIGQUIT,
				syscall.SIGTERM,
			)

			ctx := newDaemon()

			_, shouldExit, err := startDaemon(ctx)
			if err != nil {
				log.Fatalf("failed to start daemon: %+v\n", err)
			}
			if shouldExit {
				return
			}

			defer stopDaemon(ctx)

			log.Print("- - - - - - - - - - - - - - -")
			log.Print("        zyncd started")
			log.Print("- - - - - - - - - - - - - - -")

			errs := make(chan error)
			go func() { errs <- server.Start() }()
			go func() { errs <- daemon.ServeSignals() }()

			err = <-errs
			if err != nil {
				log.Fatalf("error encountered: %+v\n", err)
			}

			log.Println("daemon terminated")
		},
	}
}
