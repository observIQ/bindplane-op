package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUnboundedChan(t *testing.T) {
	unbounded := NewUnboundedChan[int](time.Millisecond)
	time.Sleep(time.Millisecond * 5)

	unbounded.In() <- 1
	unbounded.In() <- 2
	unbounded.In() <- 3
	unbounded.Close()

	x := <-unbounded.Out()
	require.Equal(t, x, 1)

	y := <-unbounded.Out()
	require.Equal(t, y, 2)

	z := <-unbounded.Out()
	require.Equal(t, z, 3)

	_, ok := <-unbounded.Out()
	require.False(t, ok)
}
