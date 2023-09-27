package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/exp/slices"
)

type Game struct {
	start              *time.Time
	config             *Config
	samplePaths        []string
	screen             tcell.Screen
	currentIndex       int
	isCurrentIndexErr  bool
	sample             *Sample
	currentSampleIndex int
	totalStrokes       int
	correctStrokes     int
	accuracy           float64
	wpm                float64
}

func newGame(c *Config) (*Game, error) {
	t := time.Now()
	g := &Game{
		start:             &t,
		config:            c,
		currentIndex:      0,
		isCurrentIndexErr: false,
	}

	g.getSamplePaths(c.entryPath)

	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	s.SetStyle(defStyle)
	g.screen = s
	if err := s.Init(); err != nil {
		return nil, err
	}

	return g, nil
}

func (g *Game) trackMetrics() {
	ticker := time.NewTicker(TrackPeriodSeconds * time.Second)

	for range ticker.C {
		minutesSince := time.Since(*g.start).Minutes()
		g.wpm = float64(g.correctStrokes) / AverageWPM / minutesSince
		g.accuracy = float64(g.correctStrokes) / float64(g.totalStrokes) * 100

		width, _ := g.screen.Size()
		g.drawText(2, 3, width, 3, defStyle, fmt.Sprintf("Accuracy: %.2f%% ", g.accuracy))
		g.drawText(2, 4, width, 4, defStyle, fmt.Sprintf("WPM: %.2f wpm", g.wpm))
	}
}

func (g *Game) updateCurrentCursorCell(cellState int) {
	width, height := g.screen.Size()
	currX, currY := getCoordinatesFromIndex(g.currentIndex, width, height)
	g.sample.state[g.currentIndex] = cellState
	g.screen.SetContent(currX, currY, rune(g.sample.text[g.currentIndex]), nil, getCellStyle(g.sample.state[g.currentIndex]))

	nextIndex := g.currentIndex + 1
	if nextIndex < len(g.sample.state) {
		g.sample.state[nextIndex] = CellCurrent
		nextX, nextY := getCoordinatesFromIndex(nextIndex, width, height)
		g.screen.SetContent(nextX, nextY, rune(g.sample.text[nextIndex]), nil, getCellStyle(g.sample.state[nextIndex]))
	}
	g.currentIndex = nextIndex
}

func (g *Game) drawSampleText() {
	startY := HeaderHeight + 2
	startX := 2
	width, height := g.screen.Size()
	row := startY
	col := startX

	for i, r := range g.sample.text {
		g.screen.SetContent(col, row, r, nil, getCellStyle(g.sample.state[i]))
		col++
		if col >= width-2 {
			row++
			col = startX
		}
		if row > height-2 {
			break
		}
	}
}

func (g *Game) drawText(startX, startY, endX, endY int, style tcell.Style, text string) {
	row := startY
	col := startX
	for _, r := range text {
		g.screen.SetContent(col, row, r, nil, style)
		col++
		if col >= endX {
			row++
			col = startX
		}
		if row > endY {
			break
		}
	}
}

func (g *Game) drawLayoyt() {
	width, height := g.screen.Size()
	for x := range make([]int, width) {
		for y := range make([]int, height) {
			if y == 0 {
				if x == 0 {
					g.screen.SetContent(x, y, tcell.RuneULCorner, nil, layoutStyle)
				} else if x == width-1 {
					g.screen.SetContent(x, y, tcell.RuneURCorner, nil, layoutStyle)
				} else {
					g.screen.SetContent(x, y, tcell.RuneHLine, nil, layoutStyle)
				}
			}
			if y == HeaderHeight {
				if x == 0 {
					g.screen.SetContent(x, y, tcell.RuneLTee, nil, layoutStyle)
				} else if x == width-1 {
					g.screen.SetContent(x, y, tcell.RuneRTee, nil, layoutStyle)
				} else {
					g.screen.SetContent(x, y, tcell.RuneHLine, nil, layoutStyle)
				}
			}
			if y == height-1 {
				if x == 0 {
					g.screen.SetContent(x, y, tcell.RuneLLCorner, nil, layoutStyle)
				} else if x == width-1 {
					g.screen.SetContent(x, y, tcell.RuneLRCorner, nil, layoutStyle)
				} else {
					g.screen.SetContent(x, y, tcell.RuneHLine, nil, layoutStyle)
				}
			}
			if x == 0 && y != 0 && y != HeaderHeight && y != height-1 {
				g.screen.SetContent(x, y, tcell.RuneVLine, nil, layoutStyle)
			}
			if x == width-1 && y != 0 && y != HeaderHeight && y != height-1 {
				g.screen.SetContent(x, y, tcell.RuneVLine, nil, layoutStyle)
			}
		}

	}
}

func (g *Game) getSamplePaths(path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatalf("Error while reading specified directory %s: %+v", path, err)
	}

	for _, e := range entries {
		if e.IsDir() {
			g.getSamplePaths(fmt.Sprintf("%s/%s", path, e.Name()))
		} else {
			shouldIgnore := true
			p := fmt.Sprintf("%s/%s", path, e.Name())
			for _, str := range g.config.ignore {
				pattern := ".*" + str + ".*"
				re := regexp.MustCompile(pattern)
				shouldIgnore = re.Match([]byte(p))
			}
			if !shouldIgnore && (len(g.config.extensions) == 0 || slices.Contains(g.config.extensions, filepath.Ext(p))) {
				g.samplePaths = append(g.samplePaths, p)
			}
		}
	}
}
func (g *Game) run() {
	g.loadSample()

	go g.trackMetrics()

	for {
		g.screen.Show()
		ev := g.screen.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventResize:
			g.screen.Sync()
			g.screen.Clear()
			g.drawLayoyt()
			g.drawSampleText()

		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				g.screen.Fini()
				g.screen.Clear()
				fmt.Println("Accuracy: ", fmt.Sprintf("%.2f%%", g.accuracy))
				fmt.Println("WPM: ", fmt.Sprintf("%.2f wpm", g.wpm))
				os.Exit(0)
				return
			} else if ev.Key() == tcell.KeyCtrlN {
				g.loadSample()
			} else {
				g.totalStrokes++
				if ev.Rune() == rune(g.sample.text[g.currentIndex]) {
					if g.isCurrentIndexErr {
						g.updateCurrentCursorCell(CellError)
					} else {
						g.updateCurrentCursorCell(CellCorrect)
					}

					if g.currentIndex >= len(g.sample.text) {
						g.loadSample()
					}

					g.isCurrentIndexErr = false
					g.correctStrokes++
					g.screen.SetContent(2, 2, ' ', nil, errStyle)
				} else {
					g.isCurrentIndexErr = true
					if unicode.IsSpace(ev.Rune()) {
						g.screen.SetContent(2, 2, tcell.RuneBlock, nil, errStyle)
					} else {
						g.screen.SetContent(2, 2, ev.Rune(), nil, errStyle)
					}
				}
			}
		}
	}

}

func (g *Game) loadSample() {
	randomIndex := rand.Intn(len(g.samplePaths) - 1)
	g.currentSampleIndex = randomIndex

	b, err := os.ReadFile(g.samplePaths[g.currentSampleIndex])
	if err != nil {
		log.Fatalf("Error while loading sample file: %+v", err)
	}
	g.screen.Clear()
	g.drawLayoyt()
	g.currentIndex = 0
	g.isCurrentIndexErr = false
	g.sample = &Sample{raw: b, text: trimWhitespace(string(b))}

	for i := range g.sample.text {
		if i == 0 {
			g.sample.state = append(g.sample.state, CellCurrent)
		} else {
			g.sample.state = append(g.sample.state, CellDefault)
		}
	}

	g.drawSampleText()
}
