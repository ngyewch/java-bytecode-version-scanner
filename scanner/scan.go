package scanner

import (
	"archive/zip"
	"context"
	"github.com/spf13/afero"
	"github.com/spf13/afero/zipfs"
	"io/fs"
	"strings"
)

type ScanContext struct {
	Parent *ScanContext
	Path   string
	FS     afero.Fs
}

var (
	RootScanContext = &ScanContext{
		Parent: nil,
		Path:   "",
		FS:     afero.NewOsFs(),
	}
)

type ProcessorFunc func(sc *ScanContext, path string) error

func (sc *ScanContext) PathComponents() []string {
	if sc.Parent == nil {
		return nil
	} else {
		return append(sc.Parent.PathComponents(), sc.Path)
	}
}

func (sc *ScanContext) Scan(ctx context.Context, path string, processorFunc ProcessorFunc) error {
	stat, err := sc.FS.Stat(path)
	if err != nil {
		return err
	}

	if stat.IsDir() {
		displayPath := path
		if !strings.HasSuffix(displayPath, "/") {
			displayPath = displayPath + "/"
		}
		subSc := &ScanContext{
			Parent: sc,
			Path:   displayPath,
			FS:     afero.NewBasePathFs(sc.FS, path),
		}
		return subSc.ScanFS(ctx, processorFunc)
	} else if strings.HasSuffix(path, ".jar") ||
		strings.HasSuffix(path, ".war") ||
		strings.HasSuffix(path, ".ear") ||
		strings.HasSuffix(path, ".zip") {
		f, err := sc.FS.Open(path)
		if err != nil {
			return err
		}
		zipReader, err := zip.NewReader(f, stat.Size())
		if err != nil {
			return err
		}
		subSc := &ScanContext{
			Parent: sc,
			Path:   path,
			FS:     zipfs.New(zipReader),
		}
		return subSc.ScanFS(ctx, processorFunc)
	} else {
		err = processorFunc(sc, path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *ScanContext) ScanFS(ctx context.Context, processorFunc ProcessorFunc) error {
	return afero.Walk(sc.FS, "", func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		return sc.Scan(ctx, path, processorFunc)
	})
}
