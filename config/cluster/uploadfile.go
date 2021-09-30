package cluster

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// UploadFile describes a file to be uploaded for the host
type UploadFile struct {
	Name            string `yaml:"name,omitempty"`
	Source          string `yaml:"src" validate:"required"`
	DestinationFile string `yaml:"dst"`
	DestinationDir  string `yaml:"dstDir"`
	PermMode        string `yaml:"perm" default:"0755"`
}

func (u UploadFile) String() string {
	if u.Name == "" {
		return u.Source
	}
	return u.Name
}

// Resolve returns a slice of UploadFiles that were found using the glob pattern or a slice
// containing the single UploadFile if it was absolute
func (u UploadFile) Resolve() ([]UploadFile, error) {
	var files []UploadFile
	if u.IsURL() {
		files = append(files, u)
		return files, nil
	}

	sources, err := filepath.Glob(u.Source)
	if err != nil {
		return nil, err
	}
	if len(sources) > 1 && u.DestinationFile != "" {
		return files, fmt.Errorf("multiple files found for '%s' but no destination directory (dstDir) set", u)
	}

	for i, s := range sources {
		name := u.Name
		if len(sources) > 1 {
			name = fmt.Sprintf("%s: %s (%d of %d)", u.Name, s, i+1, len(sources))
		}

		files = append(files, UploadFile{
			Name:            name,
			Source:          s,
			DestinationDir:  u.DestinationDir,
			DestinationFile: u.DestinationFile,
			PermMode:        u.PermMode,
		})
	}

	return files, nil
}

// IsURL returns true if the source is a URL
func (u UploadFile) IsURL() bool {
	return strings.Contains(u.Source, "://")
}

// Destination returns the target path and filename or an error if one couldn't be determined
func (u UploadFile) Destination() (string, string, error) {
	if u.DestinationDir == "" {
		if u.DestinationFile == "" {
			return "", "", fmt.Errorf("no destination set for file %s", u)
		}
		dir, fn := path.Split(u.DestinationFile)
		if dir == "" || fn == "" {
			return "", "", fmt.Errorf("destination directory not set for %s and destination is not absolute", u)
		}
		return dir, fn, nil
	}

	if u.DestinationFile != "" {
		return u.DestinationDir, u.DestinationFile, nil
	}

	return u.DestinationDir, path.Base(u.Source), nil
}
