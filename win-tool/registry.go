//go:build windows

package win_tool

import (
	"errors"
	"os/exec"
	"strings"
)

func QueryUserEnvironmentPath() ([]string, error) {
	pathCmd := exec.Command("reg", "query", "HKEY_CURRENT_USER\\Environment", "/v", "Path")
	pathByte, err := pathCmd.Output()
	if err != nil {
		return nil, err
	}
	nowPath := strings.TrimSpace(string(pathByte))
	keyWords := "Path    REG_SZ"
	if !strings.Contains(nowPath, keyWords) {
		return nil, errors.New("query path failed")
	}
	nowPath = strings.TrimSpace(nowPath[strings.LastIndex(nowPath, keyWords)+len(keyWords):])
	nowPathSlice := strings.Split(nowPath, ";")
	for i := 0; i < len(nowPathSlice); i++ {
		nowPathSlice[i] = strings.TrimSpace(nowPathSlice[i])
		if nowPathSlice[i] == "" {
			nowPathSlice = append(nowPathSlice[:i], nowPathSlice[i+1:]...)
			i--
		}
	}
	return nowPathSlice, nil
}

func QuerySystemEnvironmentPath() ([]string, error) {
	pathCmd := exec.Command("reg", "query", "HKEY_LOCAL_MACHINE\\SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Environment", "/v", "Path")
	pathByte, err := pathCmd.Output()
	if err != nil {
		return nil, err
	}
	nowPath := strings.TrimSpace(string(pathByte))
	keyWords := "Path    REG_SZ"
	if !strings.Contains(nowPath, keyWords) {
		return nil, errors.New("query path failed")
	}
	nowPath = strings.TrimSpace(nowPath[strings.LastIndex(nowPath, keyWords)+len(keyWords):])
	nowPathSlice := strings.Split(nowPath, ";")
	for i := 0; i < len(nowPathSlice); i++ {
		nowPathSlice[i] = strings.TrimSpace(nowPathSlice[i])
		if nowPathSlice[i] == "" {
			nowPathSlice = append(nowPathSlice[:i], nowPathSlice[i+1:]...)
			i--
		}
	}
	return nowPathSlice, nil
}
