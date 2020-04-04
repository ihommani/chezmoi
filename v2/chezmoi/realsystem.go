package chezmoi

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/bmatcuk/doublestar"
	"github.com/google/renameio"
	vfs "github.com/twpayne/go-vfs"
)

// An RealSystem is a System that writes to a filesystem and executes scripts.
type RealSystem struct {
	vfs.FS
	PersistentState
	devCache     map[string]uint // devCache maps directories to device numbers.
	tempDirCache map[uint]string // tempDir maps device numbers to renameio temporary directories.
}

// NewRealSystem returns a System that acts on fs.
func NewRealSystem(fs vfs.FS, persistentState PersistentState) *RealSystem {
	return &RealSystem{
		FS:              fs,
		PersistentState: persistentState,
		devCache:        make(map[string]uint),
		tempDirCache:    make(map[uint]string),
	}
}

// Glob implements System.Glob.
func (s *RealSystem) Glob(pattern string) ([]string, error) {
	return doublestar.GlobOS(s, pattern)
}

// IdempotentCmdOutput implements System.IdempotentCmdOutput.
func (s *RealSystem) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	return cmd.Output()
}

// PathSeparator implements doublestar.OS.PathSeparator.
func (s *RealSystem) PathSeparator() rune {
	return pathSeparatorRune
}

// RunScript implements System.RunScript.
func (s *RealSystem) RunScript(name string, data []byte) error {
	// Write the temporary script file. Put the randomness at the front of the
	// filename to preserve any file extension for Windows scripts.
	f, err := ioutil.TempFile("", "*."+path.Base(name))
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(f.Name())
	}()

	// Make the script private before writing it in case it contains any
	// secrets.
	if err := f.Chmod(0o700); err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err != nil {
		return err
	} else if err1 != nil {
		return err1
	}

	// Run the temporary script file.
	//nolint:gosec
	c := exec.Command(f.Name())
	// c.Dir = path.Join(applyOptions.DestDir, filepath.Dir(s.targetName)) // FIXME
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// WriteSymlink implements System.WriteSymlink.
func (s *RealSystem) WriteSymlink(oldname, newname string) error {
	// Special case: if writing to the real filesystem, use
	// github.com/google/renameio.
	if s.FS == vfs.OSFS {
		return renameio.Symlink(oldname, newname)
	}
	if err := s.FS.RemoveAll(newname); err != nil && !os.IsNotExist(err) {
		return err
	}
	return s.FS.Symlink(oldname, newname)
}

// WriteFile is like ioutil.WriteFile but always sets perm before writing data.
// ioutil.WriteFile only sets the permissions when creating a new file. We need
// to ensure permissions, so we use our own implementation.
func WriteFile(fs vfs.FS, filename string, data []byte, perm os.FileMode) error {
	// Create a new file, or truncate any existing one.
	f, err := fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}

	// From now on, we continue to the end of the function to ensure that
	// f.Close() gets called so we don't leak any file descriptors.

	// Set permissions after truncation but before writing any data, in case the
	// file contained private data before, but before writing the new contents,
	// in case the contents contain private data after.
	err = f.Chmod(perm)

	// If everything is OK so far, write the data.
	if err == nil {
		_, err = f.Write(data)
	}

	// Always call f.Close(), and overwrite the error if so far there is none.
	if err1 := f.Close(); err == nil {
		err = err1
	}

	// Return the first error encountered.
	return err
}
