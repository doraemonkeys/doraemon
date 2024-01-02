package doraemon

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// Deprecated
//
// 不等待执行完毕就返回(go程序不能立即退出),如果params中有转义字符需要自己处理,
// dir为cmd命令执行的位置,传入空值则为默认路径。
// params中的命令会用空格分隔，一次性提交给cmd。
func Cmd_NoWait(dir string, params []string) (cmd *exec.Cmd, err error) {
	cmd = exec.Command("cmd")
	cmd_in := bytes.NewBuffer(nil)
	cmd.Stdin = cmd_in
	if dir != "" {
		cmd.Dir = dir
	}
	command := ""
	for i := 0; i < len(params); i++ {
		command = command + params[i]
		if i != len(params)-1 {
			command += " "
		}
	}
	command = Utf8ToGbk(command)
	cmd_in.WriteString(command + "\n")
	err = cmd.Start() //不等待执行完毕就返回
	if err != nil {
		return cmd, err
	}
	//等待cmd已经读取指令
	for cmd_in.Len() != 0 {
		time.Sleep(time.Microsecond * 10)
	}
	return cmd, nil
}

// Deprecated
//
// 等待执行完毕才返回,
// dir为cmd命令执行的位置,传入空值则为默认路径。
// params中的命令会用空格分隔，一次性提交给cmd。
func Cmd(dir string, params []string) (string, error) {
	cmd := exec.Command("cmd")
	cmd_in := bytes.NewBuffer(nil)
	cmd.Stdin = cmd_in
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	if dir != "" {
		cmd.Dir = dir
	}
	command := ""
	for i := 0; i < len(params); i++ {
		command = command + params[i]
		if i != len(params)-1 {
			command += " "
		}
	}
	command = Utf8ToGbk(command)
	cmd_in.WriteString(command + "\n")
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	output := cmd_out.Bytes()
	return string(GbkToUtf8(output)), nil
}

// Deprecated
//
// 等待执行完毕才返回,不反回输出。
// dir为cmd命令执行的位置,传入空值则为默认路径。
// params中的命令会用空格分隔，一次性提交给cmd。
func CmdNoOutput(dir string, params []string) error {
	cmd := exec.Command("cmd")
	cmd_in := bytes.NewBuffer(nil)
	cmd.Stdin = cmd_in
	if dir != "" {
		cmd.Dir = dir
	}
	command := ""
	for i := 0; i < len(params); i++ {
		command = command + params[i]
		if i != len(params)-1 {
			command += " "
		}
	}
	command = Utf8ToGbk(command)
	cmd_in.WriteString(command + "\n")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// Press Enter Key to Continue with Timeout，超时则退出程序
func PressEnterKeyToContinueWithTimeout(timeout time.Duration) {
	ch := make(chan struct{}, 1)

	go func() {
		input := bufio.NewReader(os.Stdin)
		input.ReadString('\n')
		ch <- struct{}{}
	}()

	select {
	case <-time.After(timeout):
		os.Exit(0)
	case <-ch:
		return
	}
}

func OpenUrl(uri string) error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "start", uri)
		return cmd.Start()
	case "darwin":
		cmd := exec.Command("open", uri)
		return cmd.Start()
	case "linux":
		cmd := exec.Command("xdg-open", uri)
		return cmd.Start()
	default:
		return fmt.Errorf("don't know how to open things on %s platform", runtime.GOOS)
	}
}
