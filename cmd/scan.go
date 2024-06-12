package cmd

import (
	"fmt"
	"github.com/ngyewch/java-bytecode-version-scanner/processor"
	"github.com/ngyewch/java-bytecode-version-scanner/scanner"
	"github.com/urfave/cli/v2"
)

var (
	ScanCmd = &cli.Command{
		Name:      "Scan",
		Usage:     "Scan",
		Args:      true,
		ArgsUsage: "(path)...",
		Action:    Scan,
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:  "max-bytecode-major-version",
				Usage: "max bytecode major version",
			},
			&cli.BoolFlag{
				Name:  "list-class-files",
				Usage: "list class files",
			},
		},
	}
)

func Scan(cCtx *cli.Context) error {
	maxBytecodeMajorVersion := cCtx.Uint("max-bytecode-major-version")
	if maxBytecodeMajorVersion > 65535 {
		return fmt.Errorf("invalid max-bytecode-major-version")
	}
	listClassFiles := cCtx.Bool("list-class-files")

	p := processor.NewProcessor()
	for _, path := range cCtx.Args().Slice() {
		err := scanner.RootScanContext.Scan(cCtx.Context, path, p.Process)
		if err != nil {
			return err
		}
	}
	p.Report(uint16(maxBytecodeMajorVersion), listClassFiles)
	return nil
}
