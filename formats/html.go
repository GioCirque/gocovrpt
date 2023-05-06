package formats

import (
	"embed"
	"fmt"
	"os"
	"text/template"

	"github.com/giocirque/gocovrpt/lib"
)

const fileExt = ".html"

var (
	//go:embed assets/*
	assets embed.FS
	//go:embed templates/*
	templates    embed.FS
	supportFiles = []string{
		"assets/highlight/highlight.min.js",
		"assets/highlight/highlightjs-line-numbers.min.js",
		"assets/highlight/highlightjs-highlight-lines.min.js",
		"assets/highlight/styles/obsidian.min.css",
		"assets/gocovrpt.min.css",
	}
)

func FormatHtml(context *lib.ReportContext) error {
	fmt.Printf("Generating HTML report for %d profiles\n\n", len(context.ReportedFiles))

	// Write out supporting files, CSS, JS, etc.
	writeSupportingFile(context.Output)

	// Build the template with helper functions
	templ, err := template.New("").Funcs(template.FuncMap{
		"first": func(a []lib.ReportedBlock) lib.ReportedBlock {
			if len(a) > 0 {
				return a[0]
			}
			return lib.ReportedBlock{}
		},
		"last": func(a []lib.ReportedBlock) lib.ReportedBlock {
			if len(a) > 0 {
				return a[len(a)-1]
			}
			return lib.ReportedBlock{}
		},
		"sourceCode": func(a lib.ReportedFile) string {
			sourceCode, err := a.GetSourceCode()
			if err != nil {
				return err.Error()
			}
			return sourceCode
		},
		"swapExt": func(value string, ext string) string {
			return lib.SwapFileExt(value, ext)
		},
	}).ParseFS(templates, "templates/*.gohtml")
	if err != nil {
		return err
	}

	// Write the file reports
	for _, rptFile := range context.ReportedFiles {
		outputFile := rptFile.WithExtension(fileExt)
		file, err := lib.MakeFile(outputFile)
		if err != nil {
			return err
		}
		defer file.Close()

		err = templ.ExecuteTemplate(file, "file.gohtml", rptFile)
		if err != nil {
			return err
		}
	}

	// Write the folder reports
	allFolder := context.GetAllFolders()
	for _, rptFolder := range allFolder {
		outputFile := rptFolder.WithExtension(fileExt)
		file, err := lib.MakeFile(outputFile)
		if err != nil {
			return err
		}
		defer file.Close()

		err = templ.ExecuteTemplate(file, "folder.gohtml", rptFolder)
		if err != nil {
			return err
		}
	}

	// Write the root report file
	noFiles := make([]*lib.ReportedFile, 0)
	rootFolder := lib.NewReportedFolder(context, context.Config.SourceDir, noFiles...)
	rootFolder.ReportedFolders = context.ReportedFolders
	rootFolder.CoveredPct = context.CoveredPct
	outputFile := lib.SwapFileExt(rootFolder.OutFilePath, fileExt)
	file, err := lib.MakeFile(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	err = templ.ExecuteTemplate(file, "folder.gohtml", rootFolder)
	if err != nil {
		return err
	}

	fmt.Printf("HTML report generated at %s\n", outputFile)

	return nil
}

func writeSupportingFile(outPath string) {
	for _, supFile := range supportFiles {
		fileOutPath := outPath + "/" + supFile
		err := lib.MakeFileDir(fileOutPath)
		if err != nil {
			lib.HandleStopError(err)
		} else {
			data, err := assets.ReadFile(supFile)
			if err != nil {
				lib.HandleStopError(err)
			}

			err = os.WriteFile(fileOutPath, data, 0644)
			if err != nil {
				lib.HandleStopError(err)
			}
		}
	}
}
