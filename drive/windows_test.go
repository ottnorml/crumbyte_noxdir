//go:build windows

package drive //nolint:testpackage

import (
	"syscall"
	"testing"
)

func TestLogicalFileSize(t *testing.T) {
	data := win32finddata1{
		FileSizeHigh: 1,
		FileSizeLow:  2,
	}

	if got, want := logicalFileSize(&data), int64(1<<32+2); got != want {
		t.Fatalf("logicalFileSize() = %d, want %d", got, want)
	}
}

func TestAllocatedFileSizeForPathReturnsCompressedFileSize(t *testing.T) {
	fakeCompressedFileSize := func(_ *uint16, high *uint32) (uintptr, syscall.Errno) {
		*high = 1

		return 2, 0
	}

	got := allocatedFileSizeForPath(`C:\sparse`, 99, fakeCompressedFileSize)
	want := int64(1<<32 + 2)

	if got != want {
		t.Fatalf("allocatedFileSizeForPath() = %d, want %d", got, want)
	}
}

func TestAllocatedFileSizeForPathFallsBackOnFailure(t *testing.T) {
	fakeCompressedFileSize := func(_ *uint16, _ *uint32) (uintptr, syscall.Errno) {
		return uintptr(^uint32(0)), syscall.Errno(5)
	}

	if got, want := allocatedFileSizeForPath(`C:\sparse`, 99, fakeCompressedFileSize), int64(99); got != want {
		t.Fatalf("allocatedFileSizeForPath() = %d, want %d", got, want)
	}
}

func TestAllocatedFileSizeForPathAllowsInvalidLowWordOnSuccess(t *testing.T) {
	fakeCompressedFileSize := func(_ *uint16, high *uint32) (uintptr, syscall.Errno) {
		*high = 1

		return uintptr(^uint32(0)), 0
	}

	want := (int64(1) << 32) + int64(^uint32(0))

	if got := allocatedFileSizeForPath(`C:\sparse`, 99, fakeCompressedFileSize); got != want {
		t.Fatalf("allocatedFileSizeForPath() = %d, want %d", got, want)
	}
}
