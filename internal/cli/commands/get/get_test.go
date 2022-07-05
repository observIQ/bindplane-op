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

package get

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/cli"
)

type testArgs struct {
	pluralCmd      string
	singularCmd    string
	expectedOutput string
}

func TestGetCommand(t *testing.T) {
	tests := []testArgs{
		{
			pluralCmd:      "agents",
			singularCmd:    "agent",
			expectedOutput: "ID\tNAME   \tVERSION\tSTATUS      \tCONNECTED\tDISCONNECTED\tLABELS \n1 \tAgent 1\t1.0.0  \tConnected   \t-        \t-           \t      \t\n2 \tAgent 2\t1.0.0  \tDisconnected\t-        \t-           \t      \t\n",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("supports both singular and plural commands for %s", test.pluralCmd), func(t *testing.T) {
			buffer := bytes.NewBufferString("")
			bindplane := setupBindPlane(buffer)
			bindplane.Config.Output = tableOutput

			cmd := Command(bindplane)
			cmd.SetOut(buffer)

			cmd.SetArgs([]string{test.pluralCmd})
			executeAndAssertOutput(t, cmd, buffer, test.expectedOutput)
			buffer.Reset()

			cmd.SetArgs([]string{test.singularCmd})
			executeAndAssertOutput(t, cmd, buffer, test.expectedOutput)
		})
	}
}

func TestGetIndividualCommand(t *testing.T) {
	var tests = []struct {
		description  string
		args         []string
		expectOutput string
	}{
		{
			description:  "get agent 1",
			args:         []string{"agent", "1"},
			expectOutput: "ID\tNAME   \tVERSION\tSTATUS   \tCONNECTED\tDISCONNECTED\tLABELS \n1 \tAgent 1\t1.0.0  \tConnected\t-        \t-           \t      \t\n",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			buffer := bytes.NewBufferString("")
			bindplane := cli.NewBindPlane(common.InitConfig(""), buffer)
			bindplane.SetClient(&mockClient{})

			cmd := Command(bindplane)
			cmd.SetOut(buffer)

			cmd.SetArgs(test.args)
			cmd.Execute()

			out, err := ioutil.ReadAll(buffer)
			require.NoError(t, err)

			assert.Equal(t, test.expectOutput, string(out))
		})
	}
}
