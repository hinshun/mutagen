package daemon

import (
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/havoc-io/mutagen/pkg/filesystem"
)

const (
	// daemonDirectoryName is the name of the daemon subdirectory within the
	// Mutagen directory.
	daemonDirectoryName = "daemon"
)

// subpath computes a subpath of the daemon subdirectory, creating the daemon
// subdirectory in the process.
func subpath(name string) (string, error) {
	// Compute the daemon root directory path and ensure it exists.
	daemonRoot, err := filesystem.Mutagen(true, daemonDirectoryName)
	if err != nil {
		return "", errors.Wrap(err, "unable to compute daemon directory")
	}

	// Compute the combined path.
	return filepath.Join(daemonRoot, name), nil
}
