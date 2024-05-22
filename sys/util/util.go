package util

import (
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
)

var ostype = runtime.GOOS

func GetProcessName() (filename string) {
	name := os.Args[0]

	if ostype == "windows" {
		if strings.Contains(name, "\\") {
			reg := regexp.MustCompile("\\\\+")
			name = reg.ReplaceAllString(name, "/")
			_, filename = path.Split(name)
			index := strings.LastIndex(filename, ".")
			return filename[:index]
		}
	}

	_, filename = path.Split(name)

	return filename
}

func GetProcessPath() (filePath string) {
	name := os.Args[0]

	if ostype == "windows" {
		if strings.Contains(name, "\\") {
			reg := regexp.MustCompile("\\\\+")
			name = reg.ReplaceAllString(name, "/")
			filePath, _ = path.Split(name)
			return filePath
		}

	}

	filePath, _ = path.Split(name)

	return filePath
}

func Dirify(dir string) string {
	if strings.LastIndex(dir, "\\") != len(dir)-1 {
		dir += "\\"
	}
	return dir
}

// StringHash Change string to uint64 hash value
func StringHash(s string) (hash uint16) {
	for _, c := range s {
		ch := uint16(c)
		hash = hash + ((hash) << 5) + ch + (ch << 7)
	}
	return
}
