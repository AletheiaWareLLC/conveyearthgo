package conveyearthgo

import (
	"aletheiaware.com/netgo"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"time"
)

var (
	Years = []string{
		"2021",
	}
	Months = []string{
		"January",
		"February",
		"March",
		"April",
		"May",
		"June",
		"July",
		"August",
		"September",
		"October",
		"November",
		"December",
	}
)

var ErrDigestNotFound = errors.New("Digest Not Found")

type DigestManager interface {
	Digest(string, string) (string, time.Time, fs.File, error)
}

func NewDigestManager(fs Filesystem) DigestManager {
	return &digestManager{
		filesystem: fs,
	}
}

type digestManager struct {
	filesystem Filesystem
}

func (m *digestManager) Digest(year, month string) (string, time.Time, fs.File, error) {
	if !isYear(year) {
		return "", time.Time{}, nil, ErrDigestNotFound
	}
	index := monthIndex(month)
	if index == 0 {
		return "", time.Time{}, nil, ErrDigestNotFound
	}
	var suffix string
	if !netgo.IsLive() {
		suffix = "-Beta"
	}
	name := fmt.Sprintf("%s-%d-Convey-Digest%s.pdf", year, index, suffix)

	file, err := m.filesystem.Open(name)
	if err != nil {
		log.Println(err)
		return "", time.Time{}, nil, ErrDigestNotFound
	}

	stat, err := file.Stat()
	if err != nil {
		log.Println(err)
		return "", time.Time{}, nil, ErrDigestNotFound
	}

	return name, stat.ModTime(), file, nil
}

func isYear(year string) bool {
	for _, y := range Years {
		if y == year {
			return true
		}
	}
	return false
}

func monthIndex(month string) int {
	for i, m := range Months {
		if m == month {
			return i + 1
		}
	}
	return 0
}
