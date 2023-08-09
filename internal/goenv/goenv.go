package goenv

import (
	"errors"
	"os/exec"
	"strings"
)

func Read(varNames []string) (map[string]string, error) {
	out, err := exec.Command("go", append([]string{"env"}, varNames...)...).CombinedOutput()
	if err != nil {
		return nil, err
	}
	return parseGoEnv(varNames, out)
}

func parseGoEnv(varNames []string, data []byte) (map[string]string, error) {
	vars := make(map[string]string)

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	for i, varName := range varNames {
		if i < len(lines) && len(lines[i]) > 0 {
			vars[varName] = lines[i]
		}
	}

	if len(vars) == 0 {
		return nil, errors.New("empty env set")
	}

	return vars, nil
}
