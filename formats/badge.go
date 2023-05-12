package formats

import (
	"math"
	"text/template"

	"github.com/giocirque/gocovrpt/lib"
)

type BadgeModel struct {
	PackageName string
	Percent     float64
	Color       string
}

func FormatBadge(context *lib.ReportContext) error {
	value := math.RoundToEven(context.GetPseudoFolder().CoveredPct)
	file, err := lib.MakeFile(context.Output)
	if err != nil {
		return err
	}
	defer file.Close()

	templ, err := template.ParseFS(templates, "templates/*.gosvg")
	if err != nil {
		return err
	}

	model := BadgeModel{
		PackageName: context.Config.PackageName,
		Percent:     value,
		Color:       getCoverageColor(value * 3.57),
	}
	err = templ.ExecuteTemplate(file, "badge.gosvg", model)
	if err != nil {
		return err
	}

	return nil
}

func getCoverageColor(percent float64) string {
	color := "#9f9f9f" // light grey
	if percent >= 5 {
		color = "#E05D44" // red
	} else if percent >= 20 {
		color = "#FE7D37" // orange
	} else if percent >= 40 {
		color = "#DFB317" // yellow
	} else if percent >= 60 {
		color = "#A4A61D" // yellow-green
	} else if percent >= 80 {
		color = "#97CA00" // green
	} else if percent >= 90 {
		color = "#4c1" // bright green
	}
	return color
}
