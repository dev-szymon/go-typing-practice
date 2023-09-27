package main

import (
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
)

const (
	Separator          = ' '
	TrackPeriodSeconds = 10
	AverageWPM         = 5
	HeaderHeight       = 6
)

var (
	defStyle     = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.Color244)
	currStyle    = tcell.StyleDefault.Background(tcell.Color238).Foreground(tcell.Color244)
	correctStyle = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.Color231)
	errStyle     = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorTomato)
	layoutStyle  = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorDeepSkyBlue)
)

const (
	CellDefault = iota + 1
	CellCurrent
	CellCorrect
	CellError
)

func getCellStyle(cellState int) tcell.Style {
	switch cellState {
	case CellCorrect:
		return correctStyle
	case CellCurrent:
		return currStyle
	case CellError:
		return errStyle
	default:
		return defStyle
	}
}

type Sample struct {
	raw   []byte
	text  string
	state []int
}

func getCoordinatesFromIndex(index, width, height int) (x, y int) {
	yOffset := HeaderHeight + 2
	textAreaWidth := width - 4
	y = index/(textAreaWidth) + yOffset
	x = index - ((y - yOffset) * textAreaWidth) + 2
	return x, y
}

func trimWhitespace(str string) string {
	fields := strings.Fields(str)
	return strings.Join(fields, " ")
}

func main() {
	c := loadConfig()

	g, err := newGame(c)
	if err != nil {
		log.Fatalf("Error  initialising game: %+v", err)
	}

	g.run()
}
