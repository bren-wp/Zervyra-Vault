//go:build !windows

package core

import "os"

func atomicReplace(src, dst string) error { return os.Rename(src,dst) }
