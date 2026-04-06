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

const maxRestoreFileSize = 100 * 1024 * 1024 // 100MB 单个文件限制

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
	sessionPath := config.ConfigDir + "/sessions"
	filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if path == sessionPath || strings.HasPrefix(path, sessionPath+string(os.PathSeparator)) {
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

	if !strings.HasSuffix(header.Filename, ".tar.gz") && !strings.HasSuffix(header.Filename, ".tar") {
		http.Redirect(w, r, "/setting?err=backup_format", 302)
		return
	}

	var tr *tar.Reader
	if strings.HasSuffix(header.Filename, ".tar.gz") {
		gzr, err := gzip.NewReader(io.LimitReader(file, maxRestoreFileSize+1024*1024))
		if err != nil {
			http.Redirect(w, r, "/setting?err=backup_format", 302)
			return
		}
		defer gzr.Close()
		tr = tar.NewReader(gzr)
	} else {
		tr = tar.NewReader(io.LimitReader(file, maxRestoreFileSize+1024*1024))
	}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Redirect(w, r, "/setting?err=backup_restore", 302)
			return
		}

		target := filepath.Clean(filepath.Join(config.DataDir, hdr.Name))
		cleanBase := filepath.Clean(config.DataDir) + string(os.PathSeparator)
		if strings.Contains(hdr.Name, "..") || strings.HasPrefix(hdr.Name, "/") {
			continue
		}
		if !strings.HasPrefix(target, cleanBase) {
			continue
		}

		realPath, err := filepath.EvalSymlinks(target)
		if err != nil {
			continue
		}
		realBase, err := filepath.EvalSymlinks(config.DataDir)
		if err != nil {
			continue
		}
		if !strings.HasPrefix(realPath, realBase) {
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, 0755)
		case tar.TypeReg:
			if hdr.Size > maxRestoreFileSize {
				continue
			}
			os.MkdirAll(filepath.Dir(target), 0755)
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				continue
			}
			_, err = io.CopyN(f, tr, maxRestoreFileSize)
			f.Close()
		}
	}

	http.Redirect(w, r, "/setting?ok=1", 302)
}
