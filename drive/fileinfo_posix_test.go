//go:build linux || darwin

package drive //nolint:testpackage

import (
	"os"
	"path/filepath"
	"testing"
	"testing/quick"

	"golang.org/x/sys/unix"
)

type testAllocator struct{}

func (testAllocator) Alloc(size uint32) ([]byte, error) {
	return make([]byte, size), nil
}

func TestAllocatedSizeFromBlocks(t *testing.T) {
	tests := []struct {
		name   string
		blocks int64
		want   int64
	}{
		{name: "empty", blocks: 0, want: 0},
		{name: "one block", blocks: 1, want: 512},
		{name: "eight blocks", blocks: 8, want: 4096},
		{name: "large file", blocks: 5546336, want: 2839724032},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := allocatedSizeFromBlocks(test.blocks); got != test.want {
				t.Fatalf("allocatedSizeFromBlocks(%d) = %d, want %d", test.blocks, got, test.want)
			}
		})
	}
}

func TestNewFileInfoUsesAllocatedSize(t *testing.T) {
	stat := unix.Stat_t{
		Mode:   unix.S_IFREG,
		Size:   1 << 30,
		Blocks: 8,
	}

	info := NewFileInfo("sparse", &stat)

	if got, want := info.Size(), int64(4096); got != want {
		t.Fatalf("Size() = %d, want %d", got, want)
	}
}

func TestNewFileInfoAllocatedSizeScalesWithBlocks(t *testing.T) {
	property := func(blocks uint32) bool {
		stat := unix.Stat_t{
			Mode:   unix.S_IFREG,
			Size:   int64(blocks) * 4096,
			Blocks: int64(blocks),
		}

		return NewFileInfo("entry", &stat).Size() == int64(blocks)*512
	}

	if err := quick.Check(property, nil); err != nil {
		t.Fatal(err)
	}
}

func TestReadDirReportsAllocatedSparseFileSize(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "sparse")

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	if err = file.Truncate(1 << 30); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}

	if err = file.Close(); err != nil {
		t.Fatal(err)
	}

	var stat unix.Stat_t
	if err = unix.Lstat(filePath, &stat); err != nil {
		t.Fatal(err)
	}

	InoFilterInstance.Reset()
	entries, err := ReadDir(testAllocator{}, dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if entry.Name() != "sparse" {
			continue
		}

		if got, want := entry.Size(), allocatedSizeFromBlocks(stat.Blocks); got != want {
			t.Fatalf("ReadDir sparse size = %d, want %d", got, want)
		}

		return
	}

	t.Fatal("sparse file not found in ReadDir results")
}
