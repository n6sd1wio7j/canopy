// Package main is the entry point for the Canopy node application.
// Canopy is a proof-of-stake blockchain framework with built-in governance.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/canopy-network/canopy/app"
	"github.com/canopy-network/canopy/lib"
	"github.com/spf13/cobra"
)

const (
	// AppName is the name of the application binary
	AppName = "canopy"
	// DefaultHomeDir is the default directory for node configuration and data
	DefaultHomeDir = ".canopy"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   AppName,
	Short: "Canopy - A proof-of-stake blockchain node",
	Long: `Canopy is a modular, proof-of-stake blockchain framework
with built-in governance and committee-based consensus.`,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Canopy node",
	Long:  `Start the Canopy node and begin participating in consensus.`,
	RunE:  runStart,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Canopy node",
	Long:  `Initialize configuration files and genesis for a new Canopy node.`,
	RunE:  runInit,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Canopy",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version %s\n", AppName, lib.Version)
	},
}

func init() {
	// Register persistent flags available to all subcommands
	rootCmd.PersistentFlags().String("home", defaultHome(), "home directory for config and data")
	// Changed default log level from "debug" to "info" to reduce noise during normal operation.
	// Use --log-level=debug explicitly when troubleshooting.
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")

	// Register subcommands
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
}

// runStart starts the Canopy node process and blocks until a shutdown signal is received.
func runStart(cmd *cobra.Command, args []string) error {
	homeDir, err := cmd.Flags().GetString("home")
	if err != nil {
		return fmt.Errorf("failed to read home flag: %w", err)
	}

	logLevel, _ := cmd.Flags().GetString("log-level")
	logger := lib.NewLogger(logLevel)

	logger.Infof("Starting %s node (home=%s)", AppName, homeDir)

	// Initialize and start the application
	canopyApp, err := app.NewApp(homeDir, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}

	if err := canopyApp.Start(); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	// Wait for interrupt or termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutdown signal received, stopping node...")
	if err := canopyApp.Stop(); err != nil {
		logger.Errorf("Error during shutdown: %v", err)
	}

	logger.Info("Node stopped gracefully")
	return nil
}

// runInit initializes a new Canopy node with default configuration.
func runInit(
