package cmd

import (
	"log"
	"os"

	"github.com/icexin/eggos/cmd/egg/generate"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate an eggos loader file in this project",
	Long: `generate the zz_load_eggos.go file in this projetc directly
in order to affect the components which are built into the unikernel.

Should this file not exist, running 'egg build' will generate a temporary one.`,
	Run: func(*cobra.Command, []string) {
		err := runGenerate()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func runGenerate() (err error) {
	wd, err := os.Getwd()
	if err != nil {
		return
	}

	g, err := generate.NewGenerator(wd)
	if err != nil {
		return
	}

	return g.Generate()
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
