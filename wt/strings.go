package wt

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"unicode/utf8"
)

// sanitizeString santizes file names. It removes emojis and
// other special characters such as :,/, etc.
func sanitizeString(str string) string {
	origLen := utf8.RuneCountInString(str)
	newLen := origLen

	for _, r := range str {
		if isSanitizable(r) {
			newLen = newLen - utf8.RuneLen(r)
		}
	}

	if origLen == newLen {
		return str
	}

	newStr := make([]rune, 0, newLen)

	for _, r := range str {
		if !isSanitizable(r) {
			newStr = append(newStr, r)
		}
	}

	return string(newStr)
}

func isSanitizable(r rune) bool {
	return isSpecial(r) || isEmoji(r)
}

// isEmoji checks if a rune is an emoji or not
// Emojis rune range from WhatsApp stickers Swift repo
// https://github.com/WhatsApp/stickers/blob/master/iOS/WAStickersThirdParty/Sticker.swift#L42-L48
func isEmoji(r rune) bool {
	switch {
	case r >= 0x1F600 && r <= 0x1F64F, // Emoticons
		r >= 0x1F300 && r <= 0x1F5FF, // Misc Symbols and Pictographs
		r >= 0x1F680 && r <= 0x1F6FF, // Transport and maps
		r >= 0x2600 && r <= 0x26FF,   // Misc symbols
		r >= 0x2700 && r <= 0x27BF,   // Dingbats
		r >= 0x1F1E6 && r <= 0x1F1FF, // Flags
		r >= 0x1F900 && r <= 0x1F9FF: // Supplemental Symbols and Pictographs
		return true
	default:
		return false
	}
}

// isSpecial checks if a rune is not allowed as character for a file
// name as defined by WeTransfer
func isSpecial(c rune) bool {
	switch c {
	case '-', '_', '.', '~':
		return false
	case '$', '&', '+', ',', '/', ':', ';', '=', '?', '@':
		return true
	}
	return false
}

// ToString ouputs a string representation of WeTransfer types.
func ToString(m interface{}) string {
	var buf bytes.Buffer
	v := reflect.ValueOf(m)
	toString(&buf, v)
	return buf.String()
}

// toString is heavily inspired by go-github's Stringify which in turn was
// heavily inspired by the goprotobuf lib.

func toString(w io.Writer, val reflect.Value) {
	if val.Kind() == reflect.Ptr && val.IsNil() {
		w.Write([]byte("<nil>"))
		return
	}

	v := reflect.Indirect(val)

	switch v.Kind() {
	case reflect.String:
		fmt.Fprintf(w, `"%s"`, v)
	case reflect.Slice:
		toStringSlice(w, v)
		return
	case reflect.Struct:
		toStringStruct(w, v)
	default:
		if v.CanInterface() {
			fmt.Fprint(w, v.Interface())
		}
	}
}

func toStringSlice(w io.Writer, v reflect.Value) {
	w.Write([]byte{'['})
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			w.Write([]byte{' '})
		}
		toString(w, v.Index(i))
	}
	w.Write([]byte{']'})
}

func toStringStruct(w io.Writer, v reflect.Value) {
	if v.Type().Name() != "" {
		w.Write([]byte(v.Type().String()))
	}

	w.Write([]byte{'{'})

	var sep bool
	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		if fv.Kind() == reflect.Ptr && fv.IsNil() {
			continue
		}
		if fv.Kind() == reflect.Slice && fv.IsNil() {
			continue
		}

		if sep {
			w.Write([]byte(", "))
		} else {
			sep = true
		}
		w.Write([]byte(v.Type().Field(i).Name))
		w.Write([]byte{':'})
		toString(w, fv)
	}

	w.Write([]byte{'}'})
}
