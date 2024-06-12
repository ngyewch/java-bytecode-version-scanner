package main

import (
	"fmt"
	"github.com/ngyewch/java-bytecode-version-scanner/cmd"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	version         string
	commit          string
	commitTimestamp string

	app = &cli.App{
		Name:      "java-bytecode-version-scanner",
		Usage:     "Java bytecode version scanner",
		Args:      true,
		ArgsUsage: "(path)...",
		Version:   version,
		Action:    cmd.Scan,
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

func main() {
	cli.VersionPrinter = func(cCtx *cli.Context) {
		var parts []string
		if version != "" {
			parts = append(parts, fmt.Sprintf("version=%s", version))
		}
		if commit != "" {
			parts = append(parts, fmt.Sprintf("commit=%s", commit))
		}
		if commitTimestamp != "" {
			formattedCommitTimestamp := func(commitTimestamp string) string {
				epochSeconds, err := strconv.ParseInt(commitTimestamp, 10, 64)
				if err != nil {
					return ""
				}
				t := time.Unix(epochSeconds, 0)
				return t.Format(time.RFC3339)
			}(commitTimestamp)
			if formattedCommitTimestamp != "" {
				parts = append(parts, fmt.Sprintf("commitTimestamp=%s", formattedCommitTimestamp))
			}
		}
		fmt.Println(strings.Join(parts, " "))
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
