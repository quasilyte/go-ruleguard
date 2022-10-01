package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"
)

const maxPoints = 2048
const (
	fieldSizeX = 4
	fieldSizeY = 4
)
const tilesAtStart = 2
const probFor2 = 0.9

type button int

const (
	_ button = iota
	up
	down
	right
	left
	quit
)

var labels = func() map[button]rune {
	m := make(map[button]rune, 4)
	m[up] = 'W'
	m[down] = 'S'
	m[right] = 'D'
	m[left] = 'A'
	return m
}()
var keybinding = func() map[rune]button {
	m := make(map[rune]button, 8)
	for b, r := range labels {
		m[r] = b
		if unicode.IsUpper(r) {
			r = unicode.ToLower(r)
		} else {
			r = unicode.ToUpper(r)
		}
		m[r] = b
	}
	m[0x03] = quit
	return m
}()

var model = struct {
	Score int
	Field [fieldSizeY][fieldSizeX]int
}{}

var view = func() *template.Template {
	maxWidth := 1
	for i := maxPoints; i >= 10; i /= 10 {
		maxWidth++
	}

	w := maxWidth + 3
	r := make([]byte, fieldSizeX*w+1)
	for i := range r {
		if i%w == 0 {
			r[i] = '+'
		} else {
			r[i] = '-'
		}
	}
	rawBorder := string(r)

	v, err := template.New("").Parse(`SCORE: {{.Score}}
{{range .Field}}
` + rawBorder + `
|{{range .}} {{if .}}{{printf "%` + strconv.Itoa(maxWidth) + `d" .}}{{else}}` +
		strings.Repeat(" ", maxWidth) + `{{end}} |{{end}}{{end}}
` + rawBorder + `

(` + string(labels[up]) + `)Up (` +
		string(labels[down]) + `)Down (` +
		string(labels[left]) + `)Left (` +
		string(labels[right]) + `)Right
`)
	check(err)
	return v
}()

func check(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func clear() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	check(c.Run())
}

func draw() {
	clear()
	check(view.Execute(os.Stdout, model))
}

func addRandTile() (full bool) {
	free := make([]*int, 0, fieldSizeX*fieldSizeY)

	for x := 0; x < fieldSizeX; x++ {
		for y := 0; y < fieldSizeY; y++ {
			if model.Field[y][x] == 0 {
				free = append(free, &model.Field[y][x])
			}
		}
	}

	val := 4
	if rand.Float64() < probFor2 {
		val = 2
	}
	*free[rand.Intn(len(free))] = val

	return len(free) == 1
}

type point struct{ x, y int }

func (p point) get() int      { return model.Field[p.y][p.x] }
func (p point) set(n int)     { model.Field[p.y][p.x] = n }
func (p point) inField() bool { return p.x >= 0 && p.y >= 0 && p.x < fieldSizeX && p.y < fieldSizeY }
func (p *point) next(n point) { p.x += n.x; p.y += n.y }

func controller(key rune) (gameOver bool) {
	b := keybinding[key]

	if b == 0 {
		return false
	}
	if b == quit {
		return true
	}

	var starts []point
	var next point

	switch b {
	case up:
		next = point{0, 1}
		starts = make([]point, fieldSizeX)
		for x := 0; x < fieldSizeX; x++ {
			starts[x] = point{x, 0}
		}
	case down:
		next = point{0, -1}
		starts = make([]point, fieldSizeX)
		for x := 0; x < fieldSizeX; x++ {
			starts[x] = point{x, fieldSizeY - 1}
		}
	case right:
		next = point{-1, 0}
		starts = make([]point, fieldSizeY)
		for y := 0; y < fieldSizeY; y++ {
			starts[y] = point{fieldSizeX - 1, y}
		}
	case left:
		next = point{1, 0}
		starts = make([]point, fieldSizeY)
		for y := 0; y < fieldSizeY; y++ {
			starts[y] = point{0, y}
		}
	}

	moved := false
	winning := false

	for _, s := range starts {
		n := s
		move := func(set int) {
			moved = true
			s.set(set)
			n.set(0)
		}
		for n.next(next); n.inField(); n.next(next) {
			if s.get() != 0 {
				if n.get() == s.get() {
					score := s.get() * 2
					model.Score += score
					winning = score >= maxPoints

					move(score)
					s.next(next)
				} else if n.get() != 0 {
					s.next(next)
					if s.get() == 0 {
						move(n.get())
					}
				}
			} else if n.get() != 0 {
				move(n.get())
			}
		}
	}

	if !moved {
		return false
	}

	lost := false
	if addRandTile() {
		lost = true
	Out:
		for x := 0; x < fieldSizeX; x++ {
			for y := 0; y < fieldSizeY; y++ {
				if (y > 0 && model.Field[y][x] == model.Field[y-1][x]) ||
					(x > 0 && model.Field[y][x] == model.Field[y][x-1]) {
					lost = false
					break Out
				}
			}
		}
	}

	draw()

	if winning {
		fmt.Println("You win!")
		return true
	}
	if lost {
		fmt.Println("Game Over")
		return true
	}

	return false
}

func rosetta2048() {
	rand.Seed(time.Now().Unix())

	for i := tilesAtStart; i > 0; i-- {
		addRandTile()
	}
	draw()

	stdin := bufio.NewReader(os.Stdin)

	readKey := func() rune {
		r, _, err := stdin.ReadRune()
		check(err)
		return r
	}

	for !controller(readKey()) {
	}
}
