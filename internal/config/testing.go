package config

import (
	"fmt"
	"os"
	"path"
)

func StubConfig(main, aliases string) func() {
	orig := ReadConfigFile
	origLoc := LocalConfigFile
	LocalConfigFile = func() string {
		return path.Join(LocalConfigDir()...)
	}
	ReadConfigFile = func(fn string) ([]byte, error) {
		switch path.Base(fn) {
		case "config.yml":
			if main == "" {
				return []byte(nil), os.ErrNotExist
			} else {
				return []byte(main), nil
			}
		case "aliases.yml":
			if aliases == "" {
				return []byte(nil), os.ErrNotExist
			} else {
				return []byte(aliases), nil
			}
		default:
			return []byte(nil), fmt.Errorf("read from unstubbed file: %q", fn)
		}
	}
	return func() {
		ReadConfigFile = orig
		LocalConfigFile = origLoc
	}
}
