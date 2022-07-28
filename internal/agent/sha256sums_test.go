package agent

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSha256Sums(t *testing.T) {
	file, err := os.ReadFile("testfiles/observiq-otel-collector-v1.4.0-SHA256SUMS")
	require.NoError(t, err)
	parsed := parseSha256Sums(file)
	require.Len(t, parsed, 12)
	require.Equal(t, "245638263d8755d1fd22abbca97686226161492301489d5616f0d9b7b7734e74", parsed.sha256Sum("observiq-otel-collector_v1.4.0_linux_amd64.deb"))
	require.Equal(t, "dc133923c89ffb8ecb77e367846652a7cb5465544a8d4099f35f6fde777c5954", parsed.sha256Sum("observiq-otel-collector_v1.4.0_linux_amd64.rpm"))
}
