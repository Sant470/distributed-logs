package log

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIndex(t *testing.T) {
	f, err := os.CreateTemp(os.TempDir(), "index_test")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	c := Config{}
	c.Segment.MaxIndexBytes = 1024
	idx, err := newIndex(f, c)
	require.NoError(t, err)
	_, _, err = idx.Read(-1)
	require.Error(t, err)
	require.Equal(t, f.Name(), idx.Name())
	entries := []struct {
		off uint32
		pos uint64
	}{
		{off: 0, pos: 0},
		{off: 1, pos: 10},
	}

	for _, want := range entries {
		err = idx.Write(want.off, want.pos)
		require.NoError(t, err)
		_, pos, err := idx.Read(int64(want.off))
		require.NoError(t, err)
		require.Equal(t, want.pos, pos)
	}
	// index and scanner should error reading past existing entries
	_, _, err = idx.Read(int64(len(entries)))
	require.Equal(t, io.EOF, err)

	// index should build it's state from the existing file
	err = idx.Close()
	require.NoError(t, err)
	f, _ = os.OpenFile(f.Name(), os.O_RDWR, 0600)
	idx, err = newIndex(f, c)
	require.NoError(t, err)
	off, pos, err := idx.Read(-1)
	require.NoError(t, err)
	require.Equal(t, uint32(1), off)
	require.Equal(t, entries[1].pos, pos)
}
