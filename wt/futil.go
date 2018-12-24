package wt

import (
	"os"
)

func fileInfo(name string) (string, int64, error) {
	info, err := os.Stat(name)
	if err != nil {
		return "", 0, err
	}
	newName := stripEmojis(info.Name())
	return newName, info.Size(), nil
}

// stripEmojis removes emojis from the string and returns a new non-emojied string.
func stripEmojis(str string) string {
	strRunes := []rune(str)
	lenStrRunes := len(strRunes)

	if lenStrRunes == 0 {
		return str
	}

	var newstr []rune

	for i := 0; i < lenStrRunes; i++ {
		chunk := string(strRunes[i])
		// Todo: consider dingbats, symbols, arrows, etc.
		if len(chunk) < 3 {
			newstr = append(newstr, strRunes[i])
		}
	}

	return string(newstr)
}
