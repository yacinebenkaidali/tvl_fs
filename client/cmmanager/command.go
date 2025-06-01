package cmmanager

import (
	"fmt"
	"strings"
)

type Command uint16

var commandNames = map[string]Command{
	"upload":   UPLOAD_CMD,
	"delete":   DELETE_CMD,
	"archive":  ARCHIVE_CMD,
	"compress": COMPRESS_CMD,
	"read":     READ_CMD,
}

func (c *Command) String() string {
	for name, val := range commandNames {
		if val == *c {
			return name
		}
	}
	return "unknown"
}

func (c *Command) Set(s string) error {
	s = strings.ToLower(s)
	if val, ok := commandNames[s]; ok {
		*c = val
		return nil
	}
	return fmt.Errorf("invalid command: %q", s)
}
