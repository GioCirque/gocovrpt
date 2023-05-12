/*
Copyright © 2023 Gio Palacino <gio@palacino.net>
This file is part of CLI application gocovrpt.
*/
package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/giocirque/gocovrpt/formats"
	"github.com/giocirque/gocovrpt/lib"
	"github.com/spf13/cobra"
	"golang.org/x/tools/cover"
)

var rootCmd = &cobra.Command{
	Use:   "gocovrpt",
	Short: "Creates code coverage reports in multiple formats.",
	Long: `
gocovrpt is a CLI application that creates code coverage reports in multiple
formats like HTML, JSON, XML, TEXT, etc. with convenience options for generating
summaries, badges, and an isolated value useful in CI/CD control.

The input file MUST always be the last argument, and can support multiples separated by a space.
`,
	Example: `  $ gocovrpt -f html -l [full|summary] -o ./coverage -i ./build/coverage.raw
  $ gocovrpt -f json -l [full|summary] -o ./coverage.json -i ./build/coverage.raw
  $ gocovrpt -f xml -l [full|summary] -o ./coverage.xml -i ./build/coverage.raw
  $ gocovrpt -f badge -o ./coverage.[svg*|png|jpg] -c #00FF00 -i ./build/coverage.raw
  $ gocovrpt -f value -o ./covered -i ./build/coverage.raw`,
	Run: runRootCommand,
}

func init() {
	sourceDir, err := os.Getwd()
	if err != nil {
		sourceDir = "."
	}

	rootCmd.Flags().StringArrayP("input", "i", []string{"./.build/coverage.raw"}, "One or more coverage.raw files to read from.")
	rootCmd.Flags().StringP("format", "f", "html", fmt.Sprintf("Report format. Available formats: %s", AllFormatsString()))
	rootCmd.Flags().StringP("level", "l", "full", fmt.Sprintf("Report level. Available levels: %s", AllLevelsString()))
	rootCmd.Flags().StringP("output", "o", "./.build/coverage", "Output file or directory. For badges, the default is ./.build/coverage.svg.")
	rootCmd.Flags().StringP("source", "s", sourceDir, "The directory containing the covered source files.")
	rootCmd.Flags().StringP("package", "p", "", "The directory containing the covered source files.")
}

func Execute() {
	err := rootCmd.Execute()
	lib.HandleStopError(err)
}

func runRootCommand(cmd *cobra.Command, args []string) {
	config, err := validateArgs(cmd, args)
	lib.HandleStopError(err)

	absSourceDir, err := filepath.Abs(config.SourceDir)
	if err != nil {
		lib.HandleStopError(lib.UnresolvablePathError(config.SourceDir))
	}
	absParentRoot, err := filepath.Abs(path.Dir(config.SourceDir))
	if err != nil {
		lib.HandleStopError(lib.UnresolvablePathError(path.Dir(config.SourceDir)))
	}

	sharedMeta := lib.ReportMeta{
		PackageName: config.PackageName,
		CommonRoot:  absSourceDir,
		ParentRoot:  absParentRoot,
	}

	context := lib.NewReportContext(config, sharedMeta, config.Level == LevelFull)
	for _, input := range config.Input {
		profiles, err := cover.ParseProfiles(input)
		if err != nil {
			lib.HandleStopError(err)
		} else {
			for _, profile := range profiles {
				context.AddProfile(profile)
			}
		}
	}
	context.UpdateCoverage()

	switch config.Format {
	case FormatHtml:
		err = formats.FormatHtml(&context)
	case FormatValue:
		err = formats.FormatValue(&context)
	case FormatBadge:
		err = formats.FormatBadge(&context)
	}

	lib.HandleStopError(err)
}

func validateArgs(cmd *cobra.Command, args []string) (lib.AppConfig, error) {
	format, err := cmd.LocalFlags().GetString("format")
	if err != nil {
		return lib.AppConfig{}, err
	}
	if !IsValidFormat(format) {
		return lib.AppConfig{}, lib.InvalidArgError("format", format, AllFormats(), lib.InvalidFormatCode)
	}

	level, err := cmd.LocalFlags().GetString("level")
	if err != nil {
		return lib.AppConfig{}, err
	}
	if !IsValidLevel(level) {
		return lib.AppConfig{}, lib.InvalidArgError("level", level, AllLevels(), lib.InvalidLevelCode)
	}

	output, err := cmd.LocalFlags().GetString("output")
	if err != nil {
		return lib.AppConfig{}, err
	} else if format == FormatBadge && !cmd.LocalFlags().Changed("output") {
		// Badge output wasn't explicitly set, so make it an SVG path.
		output += ".svg"
	}

	input, err := cmd.LocalFlags().GetStringArray("input")
	if err != nil {
		return lib.AppConfig{}, err
	}

	sourceDir, err := cmd.LocalFlags().GetString("source")
	if err != nil {
		return lib.AppConfig{}, err
	}

	packageName, err := cmd.LocalFlags().GetString("package")
	if err != nil {
		return lib.AppConfig{}, err
	}
	if packageName == "" {
		fullSourcePath, _ := filepath.Abs(sourceDir)
		packageName = path.Base(path.Dir(fullSourcePath))
	}

	return lib.AppConfig{
		Format:      format,
		Level:       level,
		Output:      output,
		Input:       input,
		SourceDir:   sourceDir,
		PackageName: packageName,
	}, nil
}
