package conveyearthgo_test

import (
	"aletheiaware.com/conveyearthgo"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func Test_ReadDigests(t *testing.T) {
	var dir string
	editions, err := conveyearthgo.ReadDigests(dir)
	assert.Error(t, os.ErrNotExist)

	dir, err = ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	editions, err = conveyearthgo.ReadDigests(dir)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(editions))

	_, err = os.Create(filepath.Join(dir, "Convey-Digest-2006-01.epub"))
	assert.NoError(t, err)

	editions, err = conveyearthgo.ReadDigests(dir)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(editions))
	assert.Equal(t, "2006-01", editions[0])
}
