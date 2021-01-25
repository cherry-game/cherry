package cherryUtils

import str "strings"

type strings struct {
}

//CutLastString 截取字符串中最后一段，以@beginChar开始,@endChar结束的字符
//@text 文本
//@beginChar 开始
func (s *strings) CutLastString(text, beginChar, endChar string) string {
	if text == "" || beginChar == "" || endChar == "" {
		return ""
	}

	textRune := []rune(text)

	beginIndex := str.LastIndex(text, beginChar)

	endIndex := str.LastIndex(text, endChar)
	if endIndex < 0 {
		endIndex = len(textRune)
	}

	return string(textRune[beginIndex+1 : endIndex])
}

func (s *strings) IsBlank(value string) bool {
	return value == ""
}
