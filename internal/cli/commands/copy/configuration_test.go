package copy

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/stretchr/testify/require"
)

func setupBindPlane(buffer *bytes.Buffer) *cli.BindPlane {
	bindplane := cli.NewBindPlane(common.InitConfig(""), buffer)
	bindplane.SetClient(&mockClient{})
	return bindplane
}

type mockClient struct {
	client.BindPlane
}

var gotArgs []any

func (mc *mockClient) CopyConfig(ctx context.Context, configName, copyName string) error {
	fmt.Println("in here!")
	gotArgs = []any{configName, copyName}
	return nil
}

func TestCopyConfigCommmand(t *testing.T) {
	out := bytes.NewBufferString("")
	bp := setupBindPlane(out)
	t.Run("errors when two arguments are not present", func(t *testing.T) {
		cmd := ConfigurationCommand(bp)
		cmd.SetArgs([]string{})
		err := cmd.Execute()

		require.Error(t, err)
	})

	t.Run("calls CopyConfig with correct args", func(t *testing.T) {
		cmd := ConfigurationCommand(bp)
		cmd.SetArgs([]string{"blah", "foo"})

		cmd.Execute()

		require.Equal(t, []any{"blah", "foo"}, gotArgs)
	})
}
