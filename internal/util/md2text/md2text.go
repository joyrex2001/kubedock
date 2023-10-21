package md2text

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

// ToText will convert given markdown to a ascii text and
// wraps the text to fit within given width.
func ToText(text string, cols int) string {
	res := ""
	scanner := bufio.NewScanner(strings.NewReader(text))
	raw := false
	render := true
	table := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "```") {
			raw = !raw
			continue
		}

		if strings.HasPrefix(line, "[skip_render_start]") {
			render = false
			continue
		}

		if strings.HasPrefix(line, "[skip_render_end]") {
			render = true
			continue
		}

		if !render {
			continue
		}

		if strings.HasPrefix(line, "|") {
			table = append(table, line)
			continue
		} else {
			if len(table) > 0 {
				res += renderTable(table)
				table = []string{}
			}
		}

		if !raw {
			line = convertHeader(line)
			res += wrapString(line, cols)
		} else {
			res += "  " + line
		}
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

// convertHeader will convert a markdown header to an
// ascii alternative.
func convertHeader(text string) string {
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

	re2 := regexp.MustCompile(`(?m)\(http[^\)]*\)`)
	text = re2.ReplaceAllString(text, "")

	return text
}

// renderTable will render a markdown to text.
func renderTable(rows []string) string {
	out := ""

	headers := strings.Split(strings.Trim(rows[0], "|"), "|")
	data := make([][]string, len(rows)-2)

	for i := 2; i < len(rows); i++ {
		data[i-2] = strings.Split(strings.Trim(rows[i], "|"), "|")
	}

	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
		for _, row := range data {
			if len(row[i]) > colWidths[i] {
				colWidths[i] = len(row[i])
			}
		}
	}

	out += renderLine(colWidths)
	out += renderRow(headers, colWidths)
	out += renderLine(colWidths)
	for _, row := range data {
		out += renderRow(row, colWidths)
	}
	out += renderLine(colWidths)

	return out
}

// renderRow will render a row within a markdown table.
func renderRow(row []string, colWidths []int) string {
	out := ""
	for i, cell := range row {
		out += fmt.Sprintf("| %-*s ", colWidths[i], cell)
	}
	out += "|\n"
	return out
}

// renderLine will render a divider line in a markdown table.
func renderLine(colWidths []int) string {
	out := ""
	for _, width := range colWidths {
		out += "+"
		for i := 0; i < width+2; i++ {
			out += "-"
		}
	}
	out += "+\n"
	return out
}
