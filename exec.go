package doraemon

import (
	"bytes"
	"os/exec"
	"time"
)

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
	cmd_in.WriteString(command + "\n")
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	output := cmd_out.Bytes()
	return string(GbkToUtf8(output)), nil
}

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
	cmd_in.WriteString(command + "\n")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
