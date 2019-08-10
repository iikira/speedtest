package xmlhelper

import (
	"encoding/xml"
	"io"
)

// UnmarshalXMLData 将 r 中的 json 格式的数据, 解析到 data
func UnmarshalXMLData(r io.Reader, data interface{}) error {
	d := xml.NewDecoder(r)
	return d.Decode(data)
}

// MarshalXMLData 将 data, 生成 json 格式的数据, 写入 w 中
func MarshalXMLData(w io.Writer, data interface{}) error {
	e := xml.NewEncoder(w)
	return e.Encode(data)
}
