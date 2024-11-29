package main

import (
	"fmt"
	"log"
	"os"

	"github.com/unkaktus/vtktar"
	"github.com/urfave/cli/v2"
)

var (
	version string
)

func main() {
	app := &cli.App{
		Name:     "vtktar",
		HelpName: "vtktar",
		Usage:    "Package and extract vtktar files",
		Authors: []*cli.Author{
			{
				Name:  "Ivan Markin",
				Email: "git@unkaktus.art",
			},
		},
		Version: version,
		Commands: []*cli.Command{
			{
				Name:  "append",
				Usage: "append VTK files into a vtktar (vtktar append destination.vtktar [file.vtk ...])",
				Action: func(cCtx *cli.Context) error {
					destFilename := cCtx.Args().First()
					filenames := cCtx.Args().Tail()
					if len(filenames) == 0 {
						return fmt.Errorf("the source files are not specified")
					}
					return vtktar.Append(destFilename, filenames)
				},
			},
			{
				Name:  "extract",
				Usage: "extract VTK files from a vtktar (vtktar extract destination_directory source.vtktar)",
				Action: func(cCtx *cli.Context) error {
					destFilename := cCtx.Args().First()
					if len(cCtx.Args().Tail()) == 0 {
						return fmt.Errorf("vtktar file is not specified")
					}
					filename := cCtx.Args().Tail()[0]
					return vtktar.Extract(destFilename, filename)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal((err))
	}

}
