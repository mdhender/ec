// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package fsck

import "os"

// Exists returns true if the path exists.
func Exists(path string) bool {
	sb, err := os.Stat(path)
	if err != nil {
		return false
	}
	return sb.IsDir() || sb.Mode().IsRegular()
}

// IsDir returns true if the path exists and is a directory.
func IsDir(path string) bool {
	sb, err := os.Stat(path)
	if err != nil {
		return false
	}
	return sb.IsDir()
}

// IsFile returns true if the path exists and is a regular file.
func IsFile(path string) bool {
	sb, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !sb.IsDir() && sb.Mode().IsRegular()
}
