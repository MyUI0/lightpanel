package handlers

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func extractArchive(fpath, destDir string) bool {
	fname := filepath.Base(fpath)
	lower := strings.ToLower(fname)

	switch {
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		return extractTarGz(fpath, destDir)
	case strings.HasSuffix(lower, ".zip"):
		return extractZip(fpath, destDir)
	case strings.HasSuffix(lower, ".tar.bz2") || strings.HasSuffix(lower, ".tbz2"):
		return extractTarBz2(fpath, destDir)
	case strings.HasSuffix(lower, ".tar"):
		return extractTar(fpath, destDir)
	}
	return false
}

func extractTarGz(fpath, destDir string) bool {
	f, err := os.Open(fpath)
	if err != nil {
		return false
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return false
	}
	defer gz.Close()

	return extractTarReader(gz, destDir)
}

func extractTarBz2(fpath, destDir string) bool {
	f, err := os.Open(fpath)
	if err != nil {
		return false
	}
	defer f.Close()

	return extractTarReader(bzip2.NewReader(f), destDir)
}

func extractTar(fpath, destDir string) bool {
	f, err := os.Open(fpath)
	if err != nil {
		return false
	}
	defer f.Close()

	return extractTarReader(f, destDir)
}

func extractTarReader(r io.Reader, destDir string) bool {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false
		}
		target := filepath.Join(destDir, filepath.FromSlash(hdr.Name))
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			continue
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			_ = os.MkdirAll(target, os.FileMode(hdr.Mode))
		case tar.TypeReg:
			_ = os.MkdirAll(filepath.Dir(target), 0755)
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				continue
			}
			if _, err := io.Copy(out, tr); err != nil {
				_ = out.Close()
				return false
			}
			_ = out.Close()
		}
	}
	return true
}

func extractZip(fpath, destDir string) bool {
	r, err := zip.OpenReader(fpath)
	if err != nil {
		return false
	}
	defer r.Close()

	for _, f := range r.File {
		target := filepath.Join(destDir, filepath.FromSlash(f.Name))
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			continue
		}
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(target, f.Mode())
			continue
		}
		_ = os.MkdirAll(filepath.Dir(target), 0755)
		rc, err := f.Open()
		if err != nil {
			continue
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			continue
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return false
		}
		out.Close()
		rc.Close()
	}
	return true
}
