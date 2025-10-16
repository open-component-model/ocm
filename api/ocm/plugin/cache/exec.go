package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds"
	"ocm.software/ocm/api/utils/accessio"
)

func Exec(execpath string, config json.RawMessage, r io.Reader, w io.Writer, args ...string) ([]byte, error) {
	if len(config) > 0 {
		args = append([]string{"-c", string(config)}, args...)
	}
	cmd := exec.CommandContext(context.TODO(), execpath, args...)
	stdout := w
	if w == nil {
		stdout = accessio.LimitBuffer(accessio.DESCRIPTOR_LIMIT)
	}

	stderr := accessio.LimitBuffer(accessio.DESCRIPTOR_LIMIT)

	cmd.Stdin = r
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		var result cmds.Error
		var cerr error
		data := strings.TrimSpace(string(stderr.Bytes()))
		if len(data) == 0 {
			cerr = errors.New(err.Error())
		} else {
			// handle implicit error output from go run
			i := strings.LastIndex(data, "\n")
			for i > 0 && (i == len(data)-1 || strings.HasPrefix(data[i+1:], "exit status")) {
				data = data[:i]
				i = strings.LastIndex(data, "\n")
			}
			if err := json.Unmarshal([]byte(data), &result); err == nil {
				cerr = errors.New(result.Error)
			} else {
				if err := json.Unmarshal([]byte(data[i+1:]), &result); err == nil {
					cerr = errors.New(result.Error)
					// TODO pass effective stderr from CLI
					data = strings.TrimSpace(data[:i])
					if len(data) > 0 {
						cerr = fmt.Errorf("%w: with stderr\n%s", cerr, data)
					}
				} else {
					cerr = fmt.Errorf("[%s]", data)
				}
			}
		}
		return nil, cerr
	}
	if l, ok := stdout.(*accessio.LimitedBuffer); ok {
		if l.Exceeded() {
			return nil, fmt.Errorf("stdout limit exceeded")
		}
		return l.Bytes(), nil
	}
	return nil, nil
}
