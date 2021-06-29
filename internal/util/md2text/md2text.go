package md2text

import (
	"bufio"
	"regexp"
	"strings"
)

// ToText will convert given markdown to a ascii text.
func ToText(text string) string {
	// good enough md to text conversion
	re1 := regexp.MustCompile("(?m)^((#+) +(.*$))")
	sub := re1.FindAllStringSubmatch(text, -1)
	for i := range sub {
		ch := ""
		n := len(sub[i][3])
		switch len(sub[i][2]) {
		case 1:
			ch = "\n" + strings.Repeat("=", n)
		case 2:
			ch = "\n" + strings.Repeat("-", n)
		}
		text = strings.ReplaceAll(text, sub[i][1], sub[i][3]+ch)
	}

	re2 := regexp.MustCompile("(?m)\n```.*$")
	re3 := regexp.MustCompile(`(?m)\(http[^\)]*\)`)
	text = re2.ReplaceAllString(text, "")
	text = re3.ReplaceAllString(text, "")

	return text
}

// Wrap will wrap given text to fit within given width.
func Wrap(text string, cols int) string {
	res := ""
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		res += wrapString(line, cols)
		res += "\n"
	}
	return res
}

// wrapString will wrap a string into multiple lines in order to make it
// fit the given maximum column width.
func wrapString(text string, cols int) string {
	res := ""
	line := ""
	for _, w := range strings.Split(text, " ") {
		if len(w)+len(line) < cols {
			if line != "" {
				line += " "
			}
			line += w
		} else {
			res += line
			line = "\n" + w
		}
	}
	res += line
	return res
}
