package conveyearthgo

import (
	"os"
	"strings"
)

const (
	DIGEST_PREFIX    = "Convey-Digest-"
	DIGEST_EXTENSION = ".epub"
)

func ReadDigests(dir string) ([]string, error) {
	// Scan dir for epubs
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var editions []string
	for _, f := range files {
		if n := f.Name(); strings.HasSuffix(n, DIGEST_EXTENSION) {
			editions = append(editions, strings.TrimSuffix(strings.TrimPrefix(n, DIGEST_PREFIX), DIGEST_EXTENSION))
		}
	}
	return editions, nil
}
