//go:build linux || darwin

package structure_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/crumbyte/noxdir/structure"
)

func TestTreeTraverseUsesAllocatedSparseFileSize(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "sparse")

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0o600)
	require.NoError(t, err)

	require.NoError(t, file.Truncate(1<<30))
	require.NoError(t, file.Close())

	var stat unix.Stat_t
	require.NoError(t, unix.Lstat(filePath, &stat))

	root := structure.NewDirEntry(dir, 0)
	tree := structure.NewTree(root)

	require.NoError(t, tree.Traverse(true))
	tree.CalculateSize()

	require.Equal(t, uint64(1), root.TotalFiles)
	require.Less(t, root.Size, int64(1<<30))
	require.Equal(t, stat.Blocks*512, root.Size)
}
