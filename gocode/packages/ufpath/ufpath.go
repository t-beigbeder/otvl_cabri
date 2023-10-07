package ufpath

import (
	"strings"
)

// Join joins any number of path elements into a single path,
// separating them with /.
func Join(elem ...string) string {
	return strings.Join(elem, "/")
}

// Split splits path immediately following the final "/",
// separating it into a directory and file name component.
// If there is no "/" in path, Split returns an empty dir
// and file set to path.
// The returned values have the property that path = dir+file.
func Split(path string) (dir, file string) {
	i := len(path) - 1
	for i >= 0 && path[i] != '/' {
		i--
	}
	return path[:i+1], path[i+1:]
}

// Ext returns the file name extension used by path.
// The extension is the suffix beginning at the final dot
// in the final element of path; it is empty if there is
// no dot.
func Ext(path string) string {
	for i := len(path) - 1; i >= 0 && path[i] != '/'; i-- {
		if path[i] == '.' {
			return path[i:]
		}
	}
	return ""
}

// Base returns the last element of path.
// Trailing path separators are removed before extracting the last element.
// If the path is empty, Base returns ".".
// If the path consists entirely of separators, Base returns a single separator.
func Base(path string) string {
	if path == "" {
		return "."
	}
	// Strip trailing slashes.
	for len(path) > 0 && path[len(path)-1] == '/' {
		path = path[0 : len(path)-1]
	}
	// Find the last element
	i := len(path) - 1
	for i >= 0 && path[i] != '/' {
		i--
	}
	if i >= 0 {
		path = path[i+1:]
	}
	// If empty now, it had only slashes.
	if path == "" {
		return "/"
	}
	return path
}

// Dir returns all but the last element of path, typically the path's directory.
// If the path is empty, Dir returns ".".
// If the path is "/", Dir returns "/".
// The returned path does not end in a separator unless it is the root directory.
func Dir(path string) string {
	//other := filepath.Dir(path)
	if path == "" {
		return "."
	}
	if path == "/" {
		return "/"
	}
	i := len(path) - 1
	for i >= 0 && path[i] != '/' {
		i--
	}
	if i <= 0 {
		return "."
	}
	return path[0:i]
}
