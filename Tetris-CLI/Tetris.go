package main

import (
	"fmt"
	"golang.org/x/term"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	WIDTH             int = 10
	HEIGHT            int = 24
	EMPTY_CELL            = "  "
	BLOCK_CELL            = "[]"
	HORIZONTAL_BORDER     = "-"
	VERTICAL_BORDER       = "|"
)

var score float64 = 0

var BOARDS map[string][HEIGHT][WIDTH]int = make(map[string][HEIGHT][WIDTH]int)

var currentBoard string = "center"
var BOARD [HEIGHT][WIDTH]int
var currentPiece [][]int
var nextPiece [][]int
var piecePosition [2]int

var clear map[string]func()
var rng *rand.Rand

var fall_speed float32 = 0.5
var userInputChan = make(chan rune, 1)

var terminalWidth int = 80
var gameWidth int = 0

func init() {
	clear = make(map[string]func())
	score = 0

	clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	BOARDS["left"] = [HEIGHT][WIDTH]int{}
	BOARDS["center"] = [HEIGHT][WIDTH]int{}
	BOARDS["right"] = [HEIGHT][WIDTH]int{}

	BOARD = BOARDS[currentBoard]

	rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	gameWidth = 3*(WIDTH*len(BLOCK_CELL)+2) + 2

	updateTerminalWidth()
}

func updateTerminalWidth() {
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		terminalWidth = width
	}
}

func centerString(s string) string {
	padding := max(0, (terminalWidth-len(s))/2)
	return strings.Repeat(" ", padding) + s
}

func CallClear() {
	if clearFunc, ok := clear[runtime.GOOS]; ok {
		clearFunc()
	} else {
		fmt.Println("Warning: Unsupported platform, cannot clear terminal screen.")
	}
}

func setNext() {
	shapes := map[string][][]int{
		"I": {{1, 1, 1, 1}},
		"O": {{1, 1}, {1, 1}},
		"T": {{0, 1, 0}, {1, 1, 1}},
		"S": {{0, 1, 1}, {1, 1, 0}},
		"Z": {{1, 1, 0}, {0, 1, 1}},
		"J": {{1, 0, 0}, {1, 1, 1}},
		"L": {{0, 0, 1}, {1, 1, 1}},
	}

	keys := make([]string, 0, len(shapes))
	for k := range shapes {
		keys = append(keys, k)
	}

	if len(keys) > 0 {
		randomKey := keys[rng.Intn(len(keys))]
		nextPiece = shapes[randomKey]
	} else {
		nextPiece = [][]int{{1}}
	}
}

func isValidState(pieceToCheck [][]int, pos [2]int) bool {
	if pieceToCheck == nil {
		return false
	}
	for rIdx, rowData := range pieceToCheck {
		for cIdx, cell := range rowData {
			if cell == 1 {
				boardRow := pos[0] + rIdx
				boardCol := pos[1] + cIdx

				if boardRow < 0 || boardRow >= HEIGHT || boardCol < 0 || boardCol >= WIDTH {
					return false
				}
				if BOARD[boardRow][boardCol] == 1 {
					return false
				}
			}
		}
	}
	return true
}

func spawnPiece() {
	currentPiece = nextPiece
	if len(currentPiece) > 0 && len(currentPiece[0]) > 0 {
		piecePosition = [2]int{0, WIDTH/2 - len(currentPiece[0])/2}
	} else {
		piecePosition = [2]int{0, WIDTH / 2}
		currentPiece = [][]int{{1}}
	}

	if !isValidState(currentPiece, piecePosition) {
		CallClear()
		drawBoard()
		os.Exit(0)
	}
	setNext()
}

func drawBoard() {
	CallClear()

	leftTempBoard := BOARDS["left"]
	centerTempBoard := BOARDS["center"]
	rightTempBoard := BOARDS["right"]

	activeTempBoard := &leftTempBoard
	if currentBoard == "center" {
		activeTempBoard = &centerTempBoard
	} else if currentBoard == "right" {
		activeTempBoard = &rightTempBoard
	}

	if currentPiece != nil {
		for rIdx, rowData := range currentPiece {
			for cIdx, cell := range rowData {
				if cell == 1 {
					boardRow := piecePosition[0] + rIdx
					boardCol := piecePosition[1] + cIdx
					if boardRow >= 0 && boardRow < HEIGHT && boardCol >= 0 && boardCol < WIDTH {
						(*activeTempBoard)[boardRow][boardCol] = 1
					}
				}
			}
		}
	}

	borderLine := strings.Repeat(HORIZONTAL_BORDER, WIDTH*len(BLOCK_CELL))
	preview := strings.Split(drawNextPiece(), "\r\n")

	fmt.Println(centerString("Tetris CLI") + "\r")
	fmt.Println(centerString(fmt.Sprintf("Score: %.0f", score)) + "\r")
	fmt.Println(centerString(fmt.Sprint("Next Piece:")) + "\r")
	for _, line := range preview {
		fmt.Println(centerString(line) + "\r")
	}
	fmt.Println(centerString("+" + borderLine + "+ " + "+" + borderLine + "+ " + "+" + borderLine + "+\r"))

	for r := range HEIGHT {
		boardLine := ""
		boardLine += VERTICAL_BORDER
		for c := range WIDTH {
			if leftTempBoard[r][c] == 1 {
				boardLine += BLOCK_CELL
			} else {
				boardLine += EMPTY_CELL
			}
		}
		boardLine += VERTICAL_BORDER + " " + VERTICAL_BORDER

		for c := range WIDTH {
			if centerTempBoard[r][c] == 1 {
				boardLine += BLOCK_CELL
			} else {
				boardLine += EMPTY_CELL
			}
		}

		boardLine += VERTICAL_BORDER + " " + VERTICAL_BORDER

		for c := range WIDTH {
			if rightTempBoard[r][c] == 1 {
				boardLine += BLOCK_CELL
			} else {
				boardLine += EMPTY_CELL
			}
		}
		boardLine += VERTICAL_BORDER + "\r"
		fmt.Println(centerString(boardLine))
	}
	fmt.Println(centerString("+" + borderLine + "+ " + "+" + borderLine + "+ " + "+" + borderLine + "+\r"))

	fmt.Println(centerString("\nControls: WASD/HJKL = Movement | F = Hard Drop | Q/E = Switch Board | X = Exit"))
}

func drawNextPiece() string {
	previewWidth := 4
	previewHeight := 2

	preview := make([][]string, previewHeight)
	for i := range preview {
		preview[i] = make([]string, previewWidth)
		for j := range preview[i] {
			preview[i][j] = EMPTY_CELL
		}
	}

	offsetX := (previewWidth - len(nextPiece[0])) / 2
	offsetY := (previewHeight - len(nextPiece)) / 2

	for y, row := range nextPiece {
		for x, cell := range row {
			if cell == 1 && y+offsetY < previewHeight && x+offsetX < previewWidth {
				preview[y+offsetY][x+offsetX] = BLOCK_CELL
			}
		}
	}

	var result strings.Builder
	result.WriteString("+" + strings.Repeat(HORIZONTAL_BORDER, previewWidth*len(BLOCK_CELL)) + "+\r\n")
	for _, row := range preview {
		result.WriteString(VERTICAL_BORDER)
		for _, cell := range row {
			result.WriteString(cell)
		}
		result.WriteString(VERTICAL_BORDER + "\r\n")
	}

	result.WriteString("+" + strings.Repeat(HORIZONTAL_BORDER, previewWidth*len(BLOCK_CELL)) + "+")
	return result.String()
}

func switchBoard(newBoard string) {
	BOARDS[currentBoard] = BOARD
	currentBoard = newBoard
	BOARD = BOARDS[currentBoard]
}

func tryMoveHorizontal(deltaX int) {
	nextPosition := [2]int{piecePosition[0], piecePosition[1] + deltaX}
	if isValidState(currentPiece, nextPosition) {
		piecePosition = nextPosition
	}
}

func tryMovePieceDown() bool {
	nextPosition := [2]int{piecePosition[0] + 1, piecePosition[1]}
	if isValidState(currentPiece, nextPosition) {
		piecePosition = nextPosition
		return true
	}
	return false
}

func tryRotate() {
	if currentPiece == nil || len(currentPiece) == 0 || len(currentPiece[0]) == 0 {
		return
	}

	originalRows := len(currentPiece)
	originalCols := len(currentPiece[0])

	rotatedPiece := make([][]int, originalCols)
	for i := range rotatedPiece {
		rotatedPiece[i] = make([]int, originalRows)
	}

	for r := range currentPiece {
		for c := range currentPiece[r] {
			rotatedPiece[c][originalRows-1-r] = currentPiece[r][c]
		}
	}

	if isValidState(rotatedPiece, piecePosition) {
		currentPiece = rotatedPiece
	}
}

func hardDrop() {
	for tryMovePieceDown() {
	}
	lockPiece()
	spawnPiece()
}

func lockPiece() {
	if currentPiece == nil {
		return
	}
	for rIdx, rowData := range currentPiece {
		for cIdx, cell := range rowData {
			if cell == 1 {
				boardRow := piecePosition[0] + rIdx
				boardCol := piecePosition[1] + cIdx
				if boardRow >= 0 && boardRow < HEIGHT && boardCol >= 0 && boardCol < WIDTH {
					BOARD[boardRow][boardCol] = 1
				}
			}
		}
	}
	clearLines()

	BOARDS[currentBoard] = BOARD
}

func clearLines() {
	if currentBoard != "center" {
		return
	}

	tempScore := 0.0
	multiplier := 1.0
	linesCleared := 0

	for r := HEIGHT - 1; r >= 0; {
		isLineFull := true
		for c := range WIDTH {
			if BOARD[r][c] == 0 {
				isLineFull = false
				break
			}
		}

		if isLineFull {
			linesCleared++
			for moveR := r; moveR > 0; moveR-- {
				copy(BOARD[moveR][:], BOARD[moveR-1][:])
			}
			BOARD[0] = [WIDTH]int{}
		} else {
			r--
		}
	}

	if linesCleared > 0 {
		if linesCleared == 1 {
			tempScore = 100
		} else if linesCleared == 2 {
			tempScore = 300
		} else if linesCleared == 3 {
			tempScore = 500
		} else {
			tempScore = 850
		}
		BOARDS[currentBoard] = BOARD

		leftMultiplier := checkAndGetMultiplierFromBoard("left")
		rightMultiplier := checkAndGetMultiplierFromBoard("right")

		multiplier = (leftMultiplier * rightMultiplier)

		tempScore *= multiplier
		score += tempScore

		switchBoard("center")
	}
}

func checkAndGetMultiplierFromBoard(boardName string) float64 {
	tempCurrentBoard := currentBoard
	switchBoard(boardName)

	linesCleared := 0
	for r := HEIGHT - 1; r >= 0; {
		isLineFull := true
		for c := range WIDTH {
			if BOARD[r][c] == 0 {
				isLineFull = false
				break
			}
		}

		if isLineFull {
			linesCleared++
			for moveR := r; moveR > 0; moveR-- {
				copy(BOARD[moveR][:], BOARD[moveR-1][:])
			}
			BOARD[0] = [WIDTH]int{}
		} else {
			r--
		}
	}

	BOARDS[boardName] = BOARD
	switchBoard(tempCurrentBoard)

	if linesCleared == 0 {
		return 1
	} else if linesCleared == 1 {
		return 1.5
	} else if linesCleared == 2 {
		return 2
	} else if linesCleared == 3 {
		return 2.5
	} else {
		return 3
	}
}

func listenForInput() {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		fmt.Fprintln(os.Stderr, "Input is not a terminal. Keyboard input disabled.")
		close(userInputChan)
		return
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set terminal to raw mode: %v\n", err)
		close(userInputChan)
		return
	}
	defer term.Restore(fd, oldState)

	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			close(userInputChan)
			return
		}
		userInputChan <- rune(buf[0])
	}
}

func main() {
	setNext()
	spawnPiece()

	go listenForInput()

	drawBoard()

	ticker := time.NewTicker(time.Duration(float32(time.Second) * fall_speed))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !tryMovePieceDown() {
				lockPiece()
				spawnPiece()
			}
			drawBoard()

		case inputKey, ok := <-userInputChan:
			if !ok {
				fmt.Println("Input channel closed. Exiting.")
				return
			}

			actionTaken := true
			switch inputKey {
			case 'a', 'A', 'h', 'H':
				tryMoveHorizontal(-1)
			case 'd', 'D', 'l', 'L':
				tryMoveHorizontal(1)
			case 's', 'S', 'j', 'J':
				if !tryMovePieceDown() {
					lockPiece()
					spawnPiece()
				}
			case 'f', 'F':
				hardDrop()
			case 'w', 'W', 'k', 'K':
				tryRotate()
			case 'q', 'Q':
				switch currentBoard {
				case "left":
					switchBoard("right")
				case "right":
					switchBoard("center")
				case "center":
					switchBoard("left")
				}
			case 'e', 'E':
				switch currentBoard {
				case "left":
					switchBoard("center")
				case "right":
					switchBoard("left")
				case "center":
					switchBoard("right")
				}
			case 'x', 'X':
				return
			default:
				actionTaken = false
			}

			if actionTaken {
				drawBoard()
			}
		}
	}
}
