package conveyearthgo

import (
	"io/ioutil"
	"strings"
)

func ReadDigests(dir string) ([]string, error) {
	// Scan dir for epubs
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var editions []string
	for _, f := range files {
		if n := f.Name(); strings.HasSuffix(n, ".epub") {
			editions = append(editions, strings.TrimSuffix(strings.TrimPrefix(n, "Convey-Digest-"), ".epub"))
		}
	}
	return editions, nil
}
