package utils

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func EnsureDirExist(name string) error {
	if !FileExists(name) {
		return os.MkdirAll(name, os.ModePerm)
	}
	return nil
}

func GzipCompressFile(srcPath, dstPath string) error {
	sf, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	df, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	writer := gzip.NewWriter(df)
	writer.Name = dstPath
	writer.ModTime = time.Now().UTC()
	_, err = io.Copy(writer, sf)
	if err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}

func Sum(i []int) int {
	sum := 0
	for _, v := range i {
		sum += v
	}
	return sum
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func CurrentUTCTime() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05 +0000")
}

func IgnoreErrWriteString(writer io.Writer, s string) {
	_, _ = io.WriteString(writer, s)
}

const (
	ColorEscape = "\033["
	Green       = "32m"
	LGreen      = "92m"
	Red         = "31m"
	ColorEnd    = ColorEscape + "0m"
	Bold        = "1"
)

const (
	CharClear     = "\x1b[H\x1b[2J"
	CharTab       = "\t"
	CharNewLine   = "\r\n"
	CharCleanLine = '\x15'
)

func WrapperString(text string, color string, meta ...bool) string {
	wrapWith := make([]string, 0)
	metaLen := len(meta)
	switch metaLen {
	case 1:
		wrapWith = append(wrapWith, Bold)
	}
	wrapWith = append(wrapWith, color)
	return fmt.Sprintf("%s%s%s%s", ColorEscape, strings.Join(wrapWith, ";"), text, ColorEnd)
}

func WrapperTitle(text string) string {
	return CharNewLine + CharTab + WrapperString(text, Red, true)
}

func WrapperWarn(text string) string {
	text += "\n\r"
	return WrapperString(text, Red)
}

func IsWordCutSetChar(b rune) bool {
	switch b {
	case ' ':
		return true
	default:
		return false
	}
}

func TrimWordCutSetChar(s string) string {
	for l := len(s); len(s) > 0; {
		s = strings.TrimSpace(s)
		// add statement to trim other char here ...
		//
		//
		if len(s) < l {
			continue
		} else {
			break
		}
	}
	return s
}

func IsStringContainsWordCutSetChar(s string) bool {
	s = TrimWordCutSetChar(s)
	return strings.Contains(s, " ") // || 此处添加其他字符contains判断
}
