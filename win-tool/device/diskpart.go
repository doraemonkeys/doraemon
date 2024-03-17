//go:build windows

package device

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/doraemonkeys/doraemon"
	"github.com/doraemonkeys/doraemon/encode"
)

// 需要管理员权限
func DiskPartListDisk() ([]DiskPartDiskInfo, error) {
	cmd := exec.Command("diskpart")
	cmd_in := bytes.NewBuffer(nil)
	cmd.Stdin = cmd_in
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	_, err := cmd_in.WriteString("list disk\n")
	if err != nil {
		return nil, err
	}
	_, err = cmd_in.WriteString("exit\n")
	if err != nil {
		return nil, err
	}
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	lines, err := doraemon.ReadTrimmedLines(cmd_out)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(lines); i++ {
		if bytes.HasPrefix(lines[i], []byte("DISKPART")) {
			lines = lines[i+1:]
			break
		}
	}
	for i := 0; i < len(lines); i++ {
		if bytes.HasPrefix(lines[i], []byte("DISKPART")) {
			lines = lines[:i]
			break
		}
	}
	var diskInfos []DiskPartDiskInfo
	var linesStr []string
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(string(encode.GbkToUtf8(lines[i])))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "---") {
			continue
		}
		linesStr = append(linesStr, line)
	}
	filter := func(s string) bool {
		if strings.HasSuffix(s, "GB") ||
			strings.HasSuffix(s, "MB") ||
			strings.HasSuffix(s, "KB") {
			return true
		}
		return false
	}
	_, records, err := doraemon.ScanFields(linesStr, 1, 1,
		map[string]func(string) bool{
			"Size": filter,
			"大小":   filter,
			"Free": filter,
			"可用":   filter,
		})
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(records); i++ {
		diskInfos = append(diskInfos, DiskPartDiskInfo{
			Number: records[i][0],
			Status: records[i][1],
			Size:   records[i][2],
			Free:   records[i][3],
			Dyn:    records[i][4],
			Gpt:    records[i][5],
		})
	}
	return diskInfos, nil
}

// 卷 ###      LTR  标签         FS     类型        大小     状态       信息
// ----------  ---  -----------  -----  ----------  -------  ---------  --------
// 卷     0     E   新加卷          NTFS   磁盘分区         931 GB  正常
// 卷     1     C   system       NTFS   磁盘分区         459 GB  正常         启动
// 卷     2     D   data         NTFS   磁盘分区         464 GB  正常
// 卷     3                      FAT32  磁盘分区         599 MB  正常         系统
// 卷     4     F                FAT32  可移动           29 GB  正常
type DiskPartVolumeInfo struct {
	// 0
	Number string
	// C
	Ltr    string
	Label  string
	Fs     string
	Type   string
	Size   string
	Status string
	Info   string
}

// 磁盘 ###  状态           大小     可用     Dyn  Gpt
// --------  -------------  -------  -------  ---  ---
// 磁盘 0    联机              931 GB  1024 KB        *
// 磁盘 1    联机              931 GB  7375 MB        *
// 磁盘 2    联机               29 GB  3072 KB
type DiskPartDiskInfo struct {
	// 磁盘 0
	Number string
	// 联机
	Status string
	Size   string
	Free   string
	Dyn    string
	Gpt    string
}

// DISKPART> detail disk
//
// Generic Masstorage USB Device
// 磁盘 ID: 00000000
// 类型   : USB
// 状态 : 联机
// 路径   : 0
// 目标 : 0
// LUN ID : 0
// 位置路径 : UNAVAILABLE
// 当前只读状态: 否
// 只读: 否
// 启动磁盘: 否
// 页面文件磁盘: 否
// 休眠文件磁盘: 否
// 故障转储磁盘: 否
// 群集磁盘  : 否
//
//	卷 ###      LTR  标签         FS     类型        大小     状态       信息
//	----------  ---  -----------  -----  ----------  -------  ---------  --------
//	卷     4     F                FAT32  可移动           29 GB  正常
type DiskPartDiskDetail struct {
	Desc    string
	details []doraemon.Pair[string, string]
	Volume  []DiskPartVolumeInfo
}

// 需要管理员权限
func DiskPartDetailDisk(diskNumber uint) (DiskPartDiskDetail, error) {
	cmd := exec.Command("diskpart")
	cmd_in := bytes.NewBuffer(nil)
	cmd.Stdin = cmd_in
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	_, err := cmd_in.WriteString(fmt.Sprintf("select disk %d\n", diskNumber))
	if err != nil {
		return DiskPartDiskDetail{}, err
	}
	_, err = cmd_in.WriteString("detail disk\n")
	if err != nil {
		return DiskPartDiskDetail{}, err
	}
	_, err = cmd_in.WriteString("exit\n")
	if err != nil {
		return DiskPartDiskDetail{}, err
	}
	err = cmd.Run()
	if err != nil {
		return DiskPartDiskDetail{}, err
	}
	lines, err := doraemon.ReadTrimmedLines(cmd_out)
	if err != nil {
		return DiskPartDiskDetail{}, err
	}
	lines = lines[:len(lines)-2]
	for i := len(lines) - 1; i >= 0; i-- {
		if bytes.HasPrefix(lines[i], []byte("DISKPART")) {
			// DISKPART> detail disk
			// Generic Masstorage USB Device
			lines = lines[i+1:]
			break
		}
	}
	var diskPartDiskDetail DiskPartDiskDetail
	diskPartDiskDetail.Desc = strings.TrimSpace(string(encode.GbkToUtf8(lines[0])))
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(string(encode.GbkToUtf8(lines[i])))
		if line == "" {
			lines = lines[i+1:]
			break
		}
		keyVal := strings.Split(line, ":")
		if len(keyVal) < 2 {
			return DiskPartDiskDetail{}, fmt.Errorf("invalid line: %s", line)
		}
		diskPartDiskDetail.details = append(diskPartDiskDetail.details,
			doraemon.Pair[string, string]{
				First:  strings.TrimSpace(keyVal[0]),
				Second: strings.TrimSpace(strings.Join(keyVal[1:], ":")),
			})
	}

	var linesStr []string
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(string(encode.GbkToUtf8(lines[i])))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "---") {
			continue
		}
		linesStr = append(linesStr, line)
	}
	filter := func(s string) bool {
		if strings.HasSuffix(s, "GB") ||
			strings.HasSuffix(s, "MB") ||
			strings.HasSuffix(s, "KB") {
			return true
		}
		return false
	}
	_, records, err := doraemon.ScanFields(linesStr, 0, 1,
		map[string]func(string) bool{
			"Size": filter,
			"大小":   filter,
		})
	if err != nil {
		return DiskPartDiskDetail{}, err
	}
	var volumes []DiskPartVolumeInfo
	for i := 0; i < len(records); i++ {
		volumes = append(volumes, DiskPartVolumeInfo{
			Number: records[i][1],
			Ltr:    records[i][2],
			Label:  records[i][3],
			Fs:     records[i][4],
			Type:   records[i][5],
			Size:   records[i][6],
			Status: records[i][7],
			Info:   records[i][8],
		})
	}
	diskPartDiskDetail.Volume = volumes
	return diskPartDiskDetail, nil
}

// 需要管理员权限
func DiskPartListVolume() ([]DiskPartVolumeInfo, error) {
	cmd := exec.Command("diskpart")
	cmd_in := bytes.NewBuffer(nil)
	cmd.Stdin = cmd_in
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	_, err := cmd_in.WriteString("list volume\n")
	if err != nil {
		return nil, err
	}
	_, err = cmd_in.WriteString("exit\n")
	if err != nil {
		return nil, err
	}
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	lines, err := doraemon.ReadTrimmedLines(cmd_out)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(lines); i++ {
		if bytes.HasPrefix(lines[i], []byte("DISKPART")) {
			lines = lines[i+1:]
			break
		}
	}
	for i := 0; i < len(lines); i++ {
		if bytes.HasPrefix(lines[i], []byte("DISKPART")) {
			lines = lines[:i]
			break
		}
	}
	var linesStr []string
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(string(encode.GbkToUtf8(lines[i])))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "---") {
			continue
		}
		linesStr = append(linesStr, line)
	}
	filter := func(s string) bool {
		if strings.HasSuffix(s, "GB") ||
			strings.HasSuffix(s, "MB") ||
			strings.HasSuffix(s, "KB") {
			return true
		}
		return false
	}
	_, records, err := doraemon.ScanFields(linesStr, 0, 1,
		map[string]func(string) bool{
			"Size": filter,
			"大小":   filter,
		})
	if err != nil {
		return nil, err
	}
	var volumes []DiskPartVolumeInfo
	for i := 0; i < len(records); i++ {
		volumes = append(volumes, DiskPartVolumeInfo{
			Number: records[i][1],
			Ltr:    records[i][2],
			Label:  records[i][3],
			Fs:     records[i][4],
			Type:   records[i][5],
			Size:   records[i][6],
			Status: records[i][7],
			Info:   records[i][8],
		})
	}
	return volumes, nil
}

// 是否只读，需要管理员权限
func DiskPartIsVolumeReadOnly(diskLetter byte) (bool, error) {
	if diskLetter < 'A' || diskLetter > 'Z' {
		return false, fmt.Errorf("invalid disk letter: %c", diskLetter)
	}
	allDisk, err := DiskPartListDisk()
	if err != nil {
		return false, err
	}
	for _, disk := range allDisk {
		diskNumber, err := strconv.Atoi(strings.Split(disk.Number, " ")[1])
		if err != nil {
			return false, fmt.Errorf("invalid disk number: %s", disk.Number)
		}
		diskDetail, err := DiskPartDetailDisk(uint(diskNumber))
		if err != nil {
			return false, err
		}
		for _, volume := range diskDetail.Volume {
			if volume.Ltr == string(diskLetter) {
				status := strings.ToLower(diskDetail.details[8].Second)
				if strings.Contains(status, "是") {
					return true, nil
				}
				if strings.Contains(status, "y") {
					return true, nil
				}
				return false, nil
			}
		}
	}
	return false, fmt.Errorf("not found volume: %c", diskLetter)
}

// 磁盘是否只读，需要管理员权限
func DiskPartIsDiskReadOnly(diskNum uint) (bool, error) {
	allDisk, err := DiskPartListDisk()
	if err != nil {
		return false, err
	}
	for _, disk := range allDisk {
		num, err := strconv.Atoi(strings.Split(disk.Number, " ")[1])
		if err != nil {
			return false, fmt.Errorf("invalid disk number: %s", disk.Number)
		}
		if uint(num) != diskNum {
			continue
		}
		diskDetail, err := DiskPartDetailDisk(uint(num))
		if err != nil {
			return false, err
		}
		status := strings.ToLower(diskDetail.details[8].Second)
		if strings.Contains(status, "是") {
			return true, nil
		}
		if strings.Contains(status, "y") {
			return true, nil
		}
		return false, nil
	}
	return false, fmt.Errorf("not found disk: %d", diskNum)
}

// 设置磁盘为只读，需要管理员权限
func DiskPartSetDiskReadOnly(diskNum uint) error {
	var err error
	cmd := exec.Command("diskpart")
	cmd_in := bytes.NewBuffer(nil)
	cmd.Stdin = cmd_in
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	// cmd.Stderr = cmd_out
	_, err = cmd_in.WriteString(fmt.Sprintf("select disk %d\n", diskNum))
	if err != nil {
		return err
	}
	_, err = cmd_in.WriteString("attributes disk set readonly\n")
	if err != nil {
		return err
	}
	_, err = cmd_in.WriteString("exit\n")
	if err != nil {
		return err
	}
	err = cmd.Run()
	if err != nil {
		return err
	}
	content := encode.GbkToUtf8(cmd_out.Bytes())
	if bytes.Contains(content, []byte("success")) ||
		bytes.Contains(content, []byte("成功")) {
		return nil
	}
	return errors.New(strings.TrimSpace(string(content)))
}

// 设置磁盘为可读写，需要管理员权限
func DiskPartSetDiskReadWrite(diskNum uint) error {
	var err error
	cmd := exec.Command("diskpart")
	cmd_in := bytes.NewBuffer(nil)
	cmd.Stdin = cmd_in
	cmd_out := bytes.NewBuffer(nil)
	cmd.Stdout = cmd_out
	_, err = cmd_in.WriteString(fmt.Sprintf("select disk %d\n", diskNum))
	if err != nil {
		return err
	}
	_, err = cmd_in.WriteString("attributes disk clear readonly\n")
	if err != nil {
		return err
	}
	_, err = cmd_in.WriteString("exit\n")
	if err != nil {
		return err
	}
	err = cmd.Run()
	if err != nil {
		return err
	}
	content := encode.GbkToUtf8(cmd_out.Bytes())
	if bytes.Contains(content, []byte("success")) ||
		bytes.Contains(content, []byte("成功")) {
		return nil
	}
	return errors.New(strings.TrimSpace(string(content)))
}

// 切换磁盘只读/可读写状态，需要管理员权限
func DiskPartSwitchDiskReadOnly(diskNum uint) error {
	var readOnly bool
	var err error
	if readOnly, err = DiskPartIsDiskReadOnly(diskNum); err != nil {
		return err
	}
	if readOnly {
		return DiskPartSetDiskReadWrite(diskNum)
	}
	return DiskPartSetDiskReadOnly(diskNum)
}

// 删除卷装载点，删除后不会在下次启动时自动装载。
// 仅限于此U盘，重新插入也不会自动装载。
// 恢复使用diskpart assign
func DiskPartRemoveVolumeMountPoint(volumeLetter byte) error {
	allVolumes, err := DiskPartListVolume()
	if err != nil {
		return err
	}
	for _, volume := range allVolumes {
		if volume.Ltr == string(volumeLetter) {
			cmd := exec.Command("diskpart")
			cmd_in := bytes.NewBuffer(nil)
			cmd.Stdin = cmd_in
			cmd_out := bytes.NewBuffer(nil)
			cmd.Stdout = cmd_out
			_, err = cmd_in.WriteString(fmt.Sprintf("select volume %s\n", volume.Number))
			if err != nil {
				return err
			}
			_, err = cmd_in.WriteString("remove\n")
			if err != nil {
				return err
			}
			_, err = cmd_in.WriteString("exit\n")
			if err != nil {
				return err
			}
			err = cmd.Run()
			if err != nil {
				return err
			}
			content := encode.GbkToUtf8(cmd_out.Bytes())
			if bytes.Contains(content, []byte("success")) ||
				bytes.Contains(content, []byte("成功")) {
				return nil
			}
			return errors.New(strings.TrimSpace(string(content)))
		}
	}
	return fmt.Errorf("not found volume: %c", volumeLetter)
}

// 分配卷装载点,volumeLetter为0则自动分配
func DiskPartAssignVolumeMountPoint(num uint, volumeLetter byte) error {
	allVolumes, err := DiskPartListVolume()
	if err != nil {
		return err
	}
	for _, volume := range allVolumes {
		if volume.Number == fmt.Sprintf("%d", num) {
			cmd := exec.Command("diskpart")
			cmd_in := bytes.NewBuffer(nil)
			cmd.Stdin = cmd_in
			cmd_out := bytes.NewBuffer(nil)
			cmd.Stdout = cmd_out
			_, err = cmd_in.WriteString(fmt.Sprintf("select volume %s\n", volume.Number))
			if err != nil {
				return err
			}
			if volumeLetter == 0 {
				_, err = cmd_in.WriteString("assign\n")
			} else {
				_, err = cmd_in.WriteString(fmt.Sprintf("assign letter=%c\n", volumeLetter))
			}
			if err != nil {
				return err
			}
			_, err = cmd_in.WriteString("exit\n")
			if err != nil {
				return err
			}
			err = cmd.Run()
			if err != nil {
				return err
			}
			content := encode.GbkToUtf8(cmd_out.Bytes())
			if bytes.Contains(content, []byte("success")) ||
				bytes.Contains(content, []byte("成功")) {
				return nil
			}
			return errors.New(strings.TrimSpace(string(content)))
		}
	}
	return fmt.Errorf("not found volume: %c", volumeLetter)
}
