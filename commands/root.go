package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/from68/pp-cli/internal/format"
	"github.com/from68/pp-cli/internal/model"
	ppxml "github.com/from68/pp-cli/internal/xml"

	"github.com/spf13/cobra"
)

type contextKey string

const clientKey contextKey = "client"
const formatKey contextKey = "format"

var (
	filePath     string
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:   "pp",
	Short: "Portfolio Performance CLI — query PP XML files from the terminal",
	Long:  "pp reads Portfolio Performance XML files and exposes data via subcommands.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		f, err := ppxml.Load(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		client, err := ppxml.Decode(f)
		if err != nil {
			return fmt.Errorf("decoding file: %w", err)
		}

		outFmt, err := format.ParseFormat(outputFormat)
		if err != nil {
			return err
		}

		ctx := context.WithValue(cmd.Context(), clientKey, client)
		ctx = context.WithValue(ctx, formatKey, outFmt)
		cmd.SetContext(ctx)
		return nil
	},
}

// SetVersion sets the version string on the root command.
func SetVersion(v string) {
	rootCmd.Version = v
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&filePath, "file", "f", "", "Path to Portfolio Performance XML file (required)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, csv, tsv")
	rootCmd.MarkPersistentFlagRequired("file")
}

// clientFromContext retrieves the decoded Client from the command context.
func clientFromContext(cmd *cobra.Command) *model.Client {
	return cmd.Context().Value(clientKey).(*model.Client)
}

// formatFromContext retrieves the output format from the command context.
func formatFromContext(cmd *cobra.Command) format.OutputFormat {
	return cmd.Context().Value(formatKey).(format.OutputFormat)
}
