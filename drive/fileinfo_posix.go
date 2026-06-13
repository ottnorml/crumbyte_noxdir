//go:build linux || darwin

package drive

func allocatedSizeFromBlocks(blocks int64) int64 {
	return blocks * 512
}
