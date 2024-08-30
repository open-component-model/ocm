package output

import (
	"fmt"
	"strings"

	. "ocm.software/ocm/api/utils/out"
)

func FormatTable(ctx Context, gap string, data [][]string) {
	columns := []int{}
	maxValue := 0
	maxTitle := 0

	formats := []string{}
	if len(data) > 1 {
		for i, f := range data[0] {
			if strings.HasPrefix(f, "-") {
				formats = append(formats, "")
				data[0][i] = f[1:]
			} else {
				formats = append(formats, "-")
			}
			if len(data[0][i]) > maxTitle {
				maxTitle = len(data[0][i])
			}
		}
	}

	for _, row := range data {
		for i, col := range row {
			l := len([]rune(col))
			if i >= len(columns) {
				columns = append(columns, l)
			} else if columns[i] < l {
				columns[i] = l
			}
			if l > maxValue {
				maxValue = l
			}
		}
	}

	if len(columns) > 2 && maxValue > 200 {
		first := []string{}
		setSep := false
		for i, row := range data {
			if i == 0 {
				first = row
			} else {
				for c, col := range row {
					if c < len(first) {
						Outf(ctx, "%s%-*s: %s\n", gap, maxTitle, first[c], col)
					} else {
						Outf(ctx, "%s%d: %s\n", gap, c, col)
					}
					setSep = true
				}
				if setSep {
					Outf(ctx, "---\n")
					setSep = false
				}
			}
		}
	} else {
		format := gap
		for i, col := range columns {
			f := "-"
			if i < len(formats) {
				f = formats[i]
			}
			if i == len(columns)-1 && f == "-" {
				format = fmt.Sprintf("%s%%s ", format)
			} else {
				format = fmt.Sprintf("%s%%%s%ds ", format, f, col)
			}
		}
		format = format[:len(format)-1] + "\n"
		for _, row := range data {
			if len(row) > 0 {
				r := []interface{}{}
				for i := 0; i < len(columns); i++ {
					if i < len(row) {
						r = append(r, row[i])
					} else {
						r = append(r, "")
					}
				}
				Outf(ctx, format, r...)
			}
		}
	}
}
