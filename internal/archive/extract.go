package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	MaxArchiveBytes   int64 = 100 * 1024 * 1024
	MaxFileBytes      int64 = 50 * 1024 * 1024
	MaxTotalExtracted int64 = 500 * 1024 * 1024
)

type ExtractError struct {
	Message string
}

func (e *ExtractError) Error() string {
	return e.Message
}

func ExtractTarGz(reader io.Reader, destDir string) (string, error) {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", err
	}

	gz, err := gzip.NewReader(io.LimitReader(reader, MaxArchiveBytes+1))
	if err != nil {
		return "", &ExtractError{Message: fmt.Sprintf("invalid gzip archive: %v", err)}
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	var total int64
	var rootPrefix string
	rootSet := false

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", &ExtractError{Message: fmt.Sprintf("read tar entry: %v", err)}
		}

		if isTarMetadataEntry(header.Typeflag) {
			if err := discardTarEntry(tr, header.Size); err != nil {
				return "", &ExtractError{Message: fmt.Sprintf("skip tar metadata entry %q: %v", header.Name, err)}
			}
			continue
		}

		if !validTarPath(header.Name) {
			return "", &ExtractError{Message: fmt.Sprintf("unsafe tar path %q", header.Name)}
		}

		name := filepath.Clean(header.Name)
		parts := strings.Split(strings.Trim(name, "/"), "/")
		if len(parts) == 0 {
			continue
		}
		if !rootSet {
			rootPrefix = parts[0]
			rootSet = true
		}

		rel := strings.TrimPrefix(strings.TrimPrefix(name, rootPrefix), "/")
		if rel == "" || rel == "." {
			if header.Typeflag == tar.TypeDir {
				continue
			}
		}

		target := filepath.Join(destDir, rel)
		if err := ensureInside(destDir, target); err != nil {
			return "", err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return "", err
			}
		case tar.TypeReg:
			if header.Size > MaxFileBytes {
				return "", &ExtractError{Message: fmt.Sprintf("file %q exceeds size limit", header.Name)}
			}
			total += header.Size
			if total > MaxTotalExtracted {
				return "", &ExtractError{Message: "total extracted size exceeds limit"}
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return "", err
			}
			if err := writeRegularFile(target, tr, header.Size, header.Mode); err != nil {
				return "", err
			}
		case tar.TypeSymlink, tar.TypeLink:
			return "", &ExtractError{Message: fmt.Sprintf("unsupported link entry %q", header.Name)}
		default:
			return "", &ExtractError{Message: fmt.Sprintf("unsupported tar entry type for %q", header.Name)}
		}
	}

	return destDir, nil
}

func isTarMetadataEntry(typeflag byte) bool {
	switch typeflag {
	case tar.TypeXGlobalHeader, tar.TypeXHeader, tar.TypeGNULongName, tar.TypeGNULongLink:
		return true
	default:
		return false
	}
}

func discardTarEntry(r io.Reader, size int64) error {
	if size <= 0 {
		return nil
	}
	_, err := io.CopyN(io.Discard, r, size)
	return err
}

func validTarPath(name string) bool {
	if filepath.IsAbs(name) {
		return false
	}
	clean := filepath.Clean(name)
	if strings.HasPrefix(clean, "..") {
		return false
	}
	return !strings.Contains(name, "\\")
}

func ensureInside(base, target string) error {
	absBase, err := filepath.Abs(base)
	if err != nil {
		return err
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return &ExtractError{Message: fmt.Sprintf("path escapes extraction directory: %v", err)}
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return &ExtractError{Message: "path escapes extraction directory"}
	}
	return nil
}

func writeRegularFile(path string, r io.Reader, size int64, mode int64) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(mode))
	if err != nil {
		return err
	}

	limited := io.LimitReader(r, MaxFileBytes+1)
	written, err := io.Copy(f, limited)
	if err != nil {
		_ = f.Close()
		return err
	}
	if written > MaxFileBytes {
		_ = f.Close()
		return &ExtractError{Message: fmt.Sprintf("file %q exceeds size limit during copy", path)}
	}
	if size >= 0 && written != size {
		_ = f.Close()
		return &ExtractError{Message: fmt.Sprintf("short read for %q", path)}
	}
	return f.Close()
}
