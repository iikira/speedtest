package bytemessage

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
)

func Smessagef(format string, a ...interface{}) []byte {
	msg := fmt.Sprintf(format, a...)
	return converter.ToBytes(msg)
}
