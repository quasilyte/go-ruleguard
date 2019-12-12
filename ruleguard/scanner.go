package ruleguard

type scanner struct {
	i     int
	lines []string
}

type textLine struct {
	s   string
	num int
}

func (sc *scanner) canScan() bool {
	return sc.i < len(sc.lines)
}

func (sc *scanner) scanLines(pred func(string) bool) []textLine {
	var lines []textLine
	for {
		if sc.i >= len(sc.lines) {
			return lines
		}
		if !pred(sc.lines[sc.i]) {
			return lines
		}
		lines = append(lines, textLine{
			s:   sc.lines[sc.i],
			num: sc.i + 1,
		})
		sc.i++
	}
}
