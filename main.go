package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/viktoriaschule/management-server/config"
	"github.com/viktoriaschule/management-server/database"
	"github.com/viktoriaschule/management-server/helper"
	"github.com/viktoriaschule/management-server/log"
	"github.com/viktoriaschule/management-server/relution"
	"github.com/viktoriaschule/management-server/rest"
)

var (
	colors bool
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&colors, "colors", true, "Add colors to log")
}

func initManagementServer() {
	if colors {
		log.Colorize()
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			fmt.Println("")
			os.Exit(1)
		}
	}()
}

var rootCmd = &cobra.Command{
	Use:   "management-server",
	Short: "Backend for the management system",
	Run: func(cmd *cobra.Command, args []string) {
		c := config.GetConfig()

		db := database.NewDatabase(c)
		db.CreateTables()

		r := relution.NewRelution(c, db)
		helper.Schedule(r.FetchDevices, time.Minute)

		rest.Serve(c, db)
	},
}

func main() {
	cobra.OnInitialize(initManagementServer)
	if err := rootCmd.Execute(); err != nil {
		log.Errorf("Command failed: %v", err)
		os.Exit(1)
	}
}
