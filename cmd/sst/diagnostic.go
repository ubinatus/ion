package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sst/ion/cmd/sst/cli"
	"github.com/sst/ion/cmd/sst/mosaic/ui"
	"github.com/sst/ion/pkg/project"
)

var CmdDiagnostic = &cli.Command{
	Name: "diagnostic",
	Description: cli.Description{
		Short: "Generates a zip of diagnostics",
		Long:  "Generates a zip of diagnostic information useful for debugging issues",
	},
	Run: func(c *cli.Cli) error {
		cfg, err := project.Discover()
		if err != nil {
			return err
		}
		workingDir := project.ResolveWorkingDir(cfg)
		logDir := project.ResolveLogDir(cfg)
		logFiles, err := os.ReadDir(logDir)
		if len(logFiles) < 3 {
			ui.Error("No logs found, run your command first before generating a diagnostic report")
			return nil
		}
		fmt.Println(ui.TEXT_DIM.Render("Generating diagnostic report from last run..."))
		zipFile, err := os.Create(filepath.Join(workingDir, "report.zip"))
		if err != nil {
			return err
		}
		defer zipFile.Close()
		archive := zip.NewWriter(zipFile)
		defer archive.Close()

		addFile := func(path string, name string) error {
			fmt.Println(ui.TEXT_DIM.Render("-  " + name))
			fileToZip, err := os.Open(path)
			if err != nil {
				return err
			}
			defer fileToZip.Close()
			info, err := fileToZip.Stat()
			if err != nil {
				return err
			}
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			header.Name = name
			header.Method = zip.Deflate
			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}
			_, err = io.Copy(writer, fileToZip)
			if err != nil {
				return err
			}
			return nil
		}

		if err != nil {
			return err
		}
		for _, file := range logFiles {
			if !file.IsDir() {
				filePath := filepath.Join(logDir, file.Name())
				err := addFile(filePath, file.Name())
				if err != nil {
					return err
				}
			}
		}
		p, err := c.InitProject()
		if err != nil {
			return err
		}
		statePath, err := p.PullState()
		if err != nil {
			return err
		}
		err = addFile(statePath, "state.json")
		if err != nil {
			return err
		}
		fmt.Println()
		ui.Success("Diagnostic report generated successfully: " + zipFile.Name())
		return nil
	},
}
