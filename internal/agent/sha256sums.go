package agent

import "strings"

type sha256sums map[string]string

// parseSha256Sums is an incomplete parser of a .sha256 or -SHA256SUMS file. Technically there are separate text and
// binary modes, but we only use the text mode. See
// https://www.gnu.org/software/coreutils/manual/html_node/md5sum-invocation.html for more details on the file format
func parseSha256Sums(contents []byte) sha256sums {
	result := sha256sums{}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		parts := strings.Split(line, " ")
		if len(parts) == 3 {
			result[parts[2]] = parts[0]
		}
	}
	return result
}

func (s sha256sums) sha256Sum(file string) string {
	return s[file]
}
