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
	var total int64
	count := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false
		}
		count++
		total += hdr.Size
		if count > 10000 || hdr.Size < 0 || total > 1024*1024*1024 {
			return false
		}
		target, pathErr := archiveTarget(destDir, hdr.Name)
		if pathErr != nil {
			return false
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return false
			}
		case tar.TypeSymlink, tar.TypeLink:
			return false
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return false
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode)&0777)
			if err != nil {
				return false
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

	var total uint64
	if len(r.File) > 10000 {
		return false
	}
	for _, f := range r.File {
		if f.UncompressedSize64 > 1024*1024*1024 || total > 1024*1024*1024-f.UncompressedSize64 {
			return false
		}
		total += f.UncompressedSize64
		if f.Mode()&os.ModeSymlink != 0 {
			return false
		}
		target, pathErr := archiveTarget(destDir, f.Name)
		if pathErr != nil {
			return false
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return false
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return false
		}
		rc, err := f.Open()
		if err != nil {
			return false
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode()&0777)
		if err != nil {
			rc.Close()
			return false
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
