package formats

import (
	"fmt"

	"github.com/giocirque/gocovrpt/lib"
)

func FormatValue(context *lib.ReportContext) error {
	value := context.GetPseudoFolder().CoveredPct
	file, err := lib.MakeFile(context.Output)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("%.2f", value))

	return nil
}
