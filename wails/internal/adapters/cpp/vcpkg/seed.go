package vcpkg

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// copySeed copies what vcpkg would download (archives of PowerShell and 7-Zip, tools/*) from the
// seed folder to its downloads folder. It never overwrites: what is already there stays, so the
// second call only looks at the names.
func copySeed(ctx context.Context, seed, downloads string) error {
	if err := os.MkdirAll(downloads, 0o755); err != nil {
		return err
	}
	return filepath.WalkDir(seed, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		relative, err := filepath.Rel(seed, path)
		if err != nil || relative == "." {
			return err
		}
		target := filepath.Join(downloads, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if _, err := os.Stat(target); err == nil {
			return nil
		}
		return copyFile(path, target)
	})
}

// copyFile writes through a temporary name, so an interrupted copy never leaves a half file that
// the next run would take as done.
func copyFile(source, target string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	temporary := target + ".part"
	out, err := os.Create(temporary)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(temporary)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(temporary)
		return err
	}
	return os.Rename(temporary, target)
}
