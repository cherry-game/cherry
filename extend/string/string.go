package cherryString

import (
	"strconv"
	str "strings"
)

//CutLastString 截取字符串中最后一段，以@beginChar开始,@endChar结束的字符
//@text 文本
//@beginChar 开始
func CutLastString(text, beginChar, endChar string) string {
	if text == "" || beginChar == "" || endChar == "" {
		return ""
	}

	textRune := []rune(text)

	beginIndex := str.LastIndex(text, beginChar)

	endIndex := str.LastIndex(text, endChar)
	if endIndex < 0 || endIndex < beginIndex {
		endIndex = len(textRune)
	}

	return string(textRune[beginIndex+1 : endIndex])
}

func IsBlank(value string) bool {
	return value == ""
}

func ToStringSlice(val []interface{}) []string {
	var result []string
	for _, item := range val {
		v, ok := item.(string)
		if ok {
			result = append(result, v)
		}
	}
	return result
}

func Int64ToString(val int64) string {
	return strconv.FormatInt(val, 10)
}
