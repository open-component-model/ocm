package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds"
)

func Exec(execpath string, config json.RawMessage, r io.Reader, w io.Writer, args ...string) ([]byte, error) {
	if len(config) > 0 {
		args = append([]string{"-c", string(config)}, args...)
	}
	cmd := exec.Command(execpath, args...)
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
		var msg string
		data := stderr.Bytes()
		if len(data) == 0 {
			msg = err.Error()
		} else {
			if err := json.Unmarshal(stderr.Bytes(), &result); err == nil {
				msg = result.Error
			} else {
				msg = fmt.Sprintf("[%s]", string(stderr.Bytes()))
			}
		}
		return nil, fmt.Errorf("%s", msg)
	}
	if l, ok := stdout.(*accessio.LimitedBuffer); ok {
		if l.Exceeded() {
			return nil, fmt.Errorf("stdout limit exceeded")
		}
		return l.Bytes(), nil
	}
	return nil, nil
}
