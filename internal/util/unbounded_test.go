// Copyright  observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
