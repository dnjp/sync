package main

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

func main() {
	var configFile string
	cmd := rootCmd(&configFile)

	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	err := cmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
