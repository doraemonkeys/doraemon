package doraemon

import (
	"reflect"
	"strings"
	"testing"
)

func TestSplitBySpaceLimit2(t *testing.T) {

	tests := []struct {
		name       string
		line       string
		spaceLimit int
		want       []string
	}{
		{name: "0", line: "Access    通用 SuperSpeed USB 集线器",
			spaceLimit: 0, want: []string{"Access", "通用", "SuperSpeed", "USB", "集线器"}},
		{name: "1", line: "Access    通用 SuperSpeed USB 集线器",
			spaceLimit: 1, want: []string{"Access", "通用 SuperSpeed USB 集线器"}},
		{name: "2", line: "Access    通用 SuperSpeed USB 集线器",
			spaceLimit: 2, want: []string{"Access", "通用 SuperSpeed USB 集线器"}},
		{name: "3", line: "Access    通用 SuperSpeed USB 集线器",
			spaceLimit: 3, want: []string{"Access", "通用 SuperSpeed USB 集线器"}},
		{name: "4", line: "Access    通用 SuperSpeed USB 集线器",
			spaceLimit: 4, want: []string{"Access    通用 SuperSpeed USB 集线器"}},
		{name: "5", line: "Access    通用 SuperSpeed USB 集线器\r\n",
			spaceLimit: 0, want: []string{"Access", "通用", "SuperSpeed", "USB", "集线器"}},
		{name: "6", line: "\r\nAccess    通用 SuperSpeed USB 集线器\n",
			spaceLimit: 0, want: []string{"Access", "通用", "SuperSpeed", "USB", "集线器"}},
		{name: "7", line: "\r\nAccess    通用 SuperSpeed USB 集线器\n",
			spaceLimit: 0, want: []string{"Access", "通用", "SuperSpeed", "USB", "集线器"}},
		{name: "8", line: "      ",
			spaceLimit: 0, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitBySpaceLimit2(tt.line, tt.spaceLimit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("case %s SplitBySpaceLimit2() = %v, want %v", tt.name, got, tt.want)
			}
			if got := SplitBySpaceLimit(tt.line, tt.spaceLimit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("case %s SplitBySpaceLimit() = %#v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestScanFields(t *testing.T) {

	tests := []struct {
		name             string
		lines            []string
		fieldSpaceLimit  int
		recordSpaceLimit int
		fieldKeyWords    map[string]func(string) bool
		wantFields       []string
		wantRecords      [][]string
		wantErr          bool
	}{
		{name: "0", lines: []string{
			"Name Access  Availability  BlockSize",
			"C:\\     3       0           4096",
			"D:\\     2                  4096",
			"E:\\             1           4096",
			"F:\\     11       1           4096",
		}, fieldSpaceLimit: 0, recordSpaceLimit: 0, fieldKeyWords: nil,
			wantFields: []string{
				"Name", "Access", "Availability", "BlockSize",
			}, wantRecords: [][]string{
				{"C:\\", "3", "0", "4096"},
				{"D:\\", "2", "", "4096"},
				{"E:\\", "", "1", "4096"},
				{"F:\\", "11", "1", "4096"},
			}, wantErr: false},
		{name: "1", lines: []string{
			"Access  Availability  BlockSize  Caption  Compressed  ConfigManagerErrorCode  ConfigManagerUserConfig  CreationClassName  Description     DeviceID  DriveType  ErrorCleared  ErrorDescription  ErrorMethodology  FileSystem  FreeSpace    InstallDate  LastErrorCode  MaximumComponentLength  MediaType  Name  NumberOfBlocks  PNPDeviceID  PowerManagementCapabilities  PowerManagementSupported  ProviderName  Purpose  QuotasDisabled  QuotasIncomplete  QuotasRebuilding  Size         Status  StatusInfo  SupportsDiskQuotas  SupportsFileBasedCompression  SystemCreationClassName  SystemName  VolumeDirty  VolumeName  VolumeSerialNumber",
			"0                                F:       FALSE                                                        Win32_LogicalDisk  Removable Disk  F:        2                                                            FAT32       31109283840                              255                                F:                                                                                                                                                                   31254904832                      FALSE               FALSE                         Win32_ComputerSystem     DORAEMON    FALSE                    04030201",
		}, fieldSpaceLimit: 0, recordSpaceLimit: 1, fieldKeyWords: nil,
			wantFields: []string{
				"Access", "Availability", "BlockSize", "Caption", "Compressed", "ConfigManagerErrorCode", "ConfigManagerUserConfig", "CreationClassName", "Description", "DeviceID", "DriveType", "ErrorCleared", "ErrorDescription", "ErrorMethodology", "FileSystem", "FreeSpace", "InstallDate", "LastErrorCode", "MaximumComponentLength", "MediaType", "Name", "NumberOfBlocks", "PNPDeviceID", "PowerManagementCapabilities", "PowerManagementSupported", "ProviderName", "Purpose", "QuotasDisabled", "QuotasIncomplete", "QuotasRebuilding", "Size", "Status", "StatusInfo", "SupportsDiskQuotas", "SupportsFileBasedCompression", "SystemCreationClassName", "SystemName", "VolumeDirty", "VolumeName", "VolumeSerialNumber",
			}, wantRecords: [][]string{
				{"0", "", "", "F:", "FALSE", "", "", "Win32_LogicalDisk", "Removable Disk", "F:", "2", "", "", "", "FAT32", "31109283840", "", "", "255", "", "F:", "", "", "", "", "", "", "", "", "", "31254904832", "", "", "FALSE", "FALSE", "Win32_ComputerSystem", "DORAEMON", "FALSE", "", "04030201"},
			}, wantErr: false},

		{name: "2", lines: []string{
			"磁盘 ###  状态           大小     可用     Dyn  Gpt",
			"磁盘 0    联机              931 GB  1024 KB        *",
			"磁盘 1    联机              931 GB  7375 MB        *",
			"磁盘 2    联机               29 GB  3072 KB",
			"磁盘 4    联机               29 GB  3072 KB         ",
		}, fieldSpaceLimit: 1, recordSpaceLimit: 1, fieldKeyWords: nil,
			wantFields: []string{
				"磁盘 ###", "状态", "大小", "可用", "Dyn", "Gpt",
			}, wantRecords: [][]string{
				{"磁盘 0", "联机", "931 GB", "1024 KB", "", "*"},
				{"磁盘 1", "联机", "931 GB", "7375 MB", "", "*"},
				{"磁盘 2", "联机", "29 GB", "3072 KB", "", ""},
				{"磁盘 4", "联机", "29 GB", "3072 KB", "", ""},
			},
			wantErr: false},

		{name: "3", lines: []string{
			"卷 ###      LTR  标签         FS     类型        大小     状态       信息",
			"卷     0     E   新加卷          NTFS   磁盘分区         931 GB  正常          ",
			"卷     1     C   system       NTFS   磁盘分区         459 GB  正常         启动",
			"卷     2     D   data         NTFS   磁盘分区         464 GB  正常",
			"卷     3                      FAT32  磁盘分区         599 MB  正常         系统",
			"卷     4     F                FAT32  可移动           29 GB  正常",
		}, fieldSpaceLimit: 0, recordSpaceLimit: 1, fieldKeyWords: map[string]func(string) bool{
			"大小": func(s string) bool {
				if strings.Contains(s, "GB") ||
					strings.Contains(s, "MB") ||
					strings.Contains(s, "KB") {
					return true
				}
				return false
			},
		}, wantFields: []string{
			"卷", "###", "LTR", "标签", "FS", "类型", "大小", "状态", "信息",
		}, wantRecords: [][]string{
			{"卷", "0", "E", "新加卷", "NTFS", "磁盘分区", "931 GB", "正常", ""},
			{"卷", "1", "C", "system", "NTFS", "磁盘分区", "459 GB", "正常", "启动"},
			{"卷", "2", "D", "data", "NTFS", "磁盘分区", "464 GB", "正常", ""},
			{"卷", "3", "", "", "FAT32", "磁盘分区", "599 MB", "正常", "系统"},
			{"卷", "4", "F", "", "FAT32", "可移动", "29 GB", "正常", ""},
		},
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFields, gotRecords, err := ScanFields(tt.lines, tt.fieldSpaceLimit, tt.recordSpaceLimit, tt.fieldKeyWords)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if len(gotFields) != len(tt.wantFields) {
				t.Errorf("ScanFields() gotFields = %v, want %v", gotFields, tt.wantFields)
				return
			}
			if len(gotRecords) != len(tt.wantRecords) {
				t.Errorf("ScanFields() gotRecords = %v, want %v", gotRecords, tt.wantRecords)
				return
			}
			for i := 0; i < len(gotFields); i++ {
				if gotFields[i] != tt.wantFields[i] {
					t.Errorf("ScanFields() gotFields[%d] = %v, want %v", i, gotFields[i], tt.wantFields[i])
					return
				}
			}
			for i := 0; i < len(gotRecords); i++ {
				for j := 0; j < len(gotRecords[i]); j++ {
					if gotRecords[i][j] != tt.wantRecords[i][j] {
						t.Errorf("ScanFields() field=%s gotRecords[%d][%d] = %v, want %v", gotFields[j], i, j, gotRecords[i][j], tt.wantRecords[i][j])
						return
					}
				}
			}
		})
	}
}
