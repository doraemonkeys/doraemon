package other

import (
	"fmt"

	"github.com/fatih/color"
)

// colors描述了后面每个字符串的颜色属性，colors与strs长度必须相同,
// 注意字符串不要忘了带上空格和换行。
func ColorPrint(attributes []color.Attribute, strs ...string) {
	for k, str := range strs {
		if attributes[k] != 0 {
			color.Set(attributes[k])
			fmt.Print(str)
			color.Unset()
		} else {
			fmt.Print(str)
		}
	}
}
