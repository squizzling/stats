package iio

import (
	"os"
)

func IsSymlink(mode os.FileMode) bool {
	// The slavish dedication to gofmt which produces the abomination
	// of "mode&os" requires this file to be isolated and buried lest
	// the rest of the project become infected with un-readable code.
	return mode&os.ModeSymlink != 0
}

func IsDir(mode os.FileMode) bool {
	return mode&os.ModeDir != 0
}
