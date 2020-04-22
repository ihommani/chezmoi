package chezmoi

import (
	"fmt"
	"os"
)

// Configuration constants.
const (
	pathSeparator    = '/'
	pathSeparatorStr = string(pathSeparator)
	ignorePrefix     = "."
)

// Configuration variables.
var (
	scriptOnceStateBucket = []byte("script")
)

// Suffixes and prefixes.
const (
	dotPrefix        = "dot_"
	emptyPrefix      = "empty_"
	encryptedPrefix  = "encrypted_"
	exactPrefix      = "exact_"
	executablePrefix = "executable_"
	oncePrefix       = "once_"
	privatePrefix    = "private_"
	runPrefix        = "run_"
	symlinkPrefix    = "symlink_"
	templateSuffix   = ".tmpl"
)

// Special file names.
const (
	chezmoiPrefix = ".chezmoi"

	ignoreName       = chezmoiPrefix + "ignore"
	removeName       = chezmoiPrefix + "remove"
	templatesDirName = chezmoiPrefix + "templates"
	versionName      = chezmoiPrefix + "version"
)

// DefaultTemplateOptions are the default template options.
var DefaultTemplateOptions = []string{"missingkey=error"}

var modeTypeNames = map[os.FileMode]string{
	0:                 "file",
	os.ModeDir:        "dir",
	os.ModeSymlink:    "symlink",
	os.ModeNamedPipe:  "named pipe",
	os.ModeSocket:     "socket",
	os.ModeDevice:     "device",
	os.ModeCharDevice: "char device",
}

func modeTypeName(mode os.FileMode) string {
	if name, ok := modeTypeNames[mode&os.ModeType]; ok {
		return name
	}
	return fmt.Sprintf("unknown (%d)", mode&os.ModeType)
}
