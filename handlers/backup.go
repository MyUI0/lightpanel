package handlers

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"lightpanel/config"
)

func backupData(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := "lightpanel_backup_" + timestamp + ".tar.gz"
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)

	gz := gzip.NewWriter(w)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	baseDir := config.DataDir
	filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(filepath.Dir(baseDir), path)
		rel = strings.ReplaceAll(rel, "\\", "/")

		if info.IsDir() {
			hdr := &tar.Header{
				Name:     rel + "/",
				Mode:     0755,
				ModTime:  info.ModTime(),
				Typeflag: tar.TypeDir,
			}
			tw.WriteHeader(hdr)
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		hdr := &tar.Header{
			Name:     rel,
			Mode:     0644,
			Size:     info.Size(),
			ModTime:  info.ModTime(),
			Typeflag: tar.TypeReg,
		}
		tw.WriteHeader(hdr)
		io.Copy(tw, f)
		return nil
	})
}

func restoreData(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}

	file, header, err := r.FormFile("backup_file")
	if err != nil {
		http.Redirect(w, r, "/setting?err=backup_file", 302)
		return
	}
	defer file.Close()

	if !strings.HasSuffix(header.Filename, ".tar.gz") {
		http.Redirect(w, r, "/setting?err=backup_format", 302)
		return
	}

	gzr, err := gzip.NewReader(file)
	if err != nil {
		http.Redirect(w, r, "/setting?err=backup_format", 302)
		return
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Redirect(w, r, "/setting?err=backup_restore", 302)
			return
		}

		target := filepath.Join(config.DataDir, hdr.Name)
		target = filepath.Clean(target)
		if !strings.HasPrefix(target, filepath.Clean(config.DataDir)) {
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, 0755)
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(target), 0755)
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				continue
			}
			io.Copy(f, tr)
			f.Close()
		}
	}

	http.Redirect(w, r, "/setting?ok=1", 302)
}
