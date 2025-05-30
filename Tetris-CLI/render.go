package main

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Renderer struct {
	game  *Game
	clear map[string]func()
}

func NewRenderer(game *Game) *Renderer {
	r := &Renderer{
		game:  game,
		clear: make(map[string]func()),
	}

	r.clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	r.clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	r.UpdateTerminalWidth()
	return r
}

func (r *Renderer) UpdateTerminalWidth() {
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		r.game.terminalWidth = width
	}
}

func (r *Renderer) ClearScreen() {
	if clearFunc, ok := r.clear[runtime.GOOS]; ok {
		clearFunc()
	} else {
		fmt.Println("Warning: Unsupported platform, cannot clear terminal screen.")
	}
}

func (r *Renderer) CenterString(s string) string {
	padding := max(0, (r.game.terminalWidth-len(s))/2)
	return strings.Repeat(" ", padding) + s
}

func (r *Renderer) DrawNextPiece() string {
	previewWidth := 4
	previewHeight := 2

	preview := make([][]string, previewHeight)
	for i := range preview {
		preview[i] = make([]string, previewWidth)
		for j := range preview[i] {
			preview[i][j] = EMPTY_CELL
		}
	}

	offsetX := (previewWidth - len(r.game.nextPiece[0])) / 2
	offsetY := (previewHeight - len(r.game.nextPiece)) / 2

	for y, row := range r.game.nextPiece {
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

func (r *Renderer) DrawBoard() {
	r.ClearScreen()

	leftTempBoard := r.game.boards["left"]
	centerTempBoard := r.game.boards["center"]
	rightTempBoard := r.game.boards["right"]

	activeTempBoard := &leftTempBoard
	if r.game.currentBoard == "center" {
		activeTempBoard = &centerTempBoard
	} else if r.game.currentBoard == "right" {
		activeTempBoard = &rightTempBoard
	}

	if r.game.currentPiece != nil {
		for rowIdx, rowData := range r.game.currentPiece {
			for cellIdx, cell := range rowData {
				if cell == 1 {
					boardRow := r.game.piecePosition[0] + rowIdx
					boardCol := r.game.piecePosition[1] + cellIdx
					if boardRow >= 0 && boardRow < HEIGHT && boardCol >= 0 && boardCol < WIDTH {
						(*activeTempBoard)[boardRow][boardCol] = 1
					}
				}
			}
		}
	}

	borderLine := strings.Repeat(HORIZONTAL_BORDER, WIDTH*len(BLOCK_CELL))
	preview := strings.Split(r.DrawNextPiece(), "\r\n")

	fmt.Println(r.CenterString("Tetris CLI") + "\r")
	fmt.Println(r.CenterString(fmt.Sprintf("Score: %.0f", r.game.score)) + "\r")
	fmt.Println(r.CenterString(fmt.Sprint("Next Piece:")) + "\r")

	for _, line := range preview {
		fmt.Println(r.CenterString(line) + "\r")
	}
	fmt.Println(r.CenterString("+" + borderLine + "+ " + "+" + borderLine + "+ " + "+" + borderLine + "+\r"))

	for row := range HEIGHT {
		boardLine := ""
		boardLine += VERTICAL_BORDER
		for c := range WIDTH {
			if leftTempBoard[row][c] == 1 {
				boardLine += BLOCK_CELL
			} else {
				boardLine += EMPTY_CELL
			}
		}
		boardLine += VERTICAL_BORDER + " " + VERTICAL_BORDER

		for c := range WIDTH {
			if centerTempBoard[row][c] == 1 {
				boardLine += BLOCK_CELL
			} else {
				boardLine += EMPTY_CELL
			}
		}

		boardLine += VERTICAL_BORDER + " " + VERTICAL_BORDER

		for c := range WIDTH {
			if rightTempBoard[row][c] == 1 {
				boardLine += BLOCK_CELL
			} else {
				boardLine += EMPTY_CELL
			}
		}
		boardLine += VERTICAL_BORDER + "\r"
		fmt.Println(r.CenterString(boardLine))
	}
	fmt.Println(r.CenterString("+" + borderLine + "+ " + "+" + borderLine + "+ " + "+" + borderLine + "+\r"))
	fmt.Println(r.CenterString("\nControls: WASD/HJKL = Movement | F = Hard Drop | Q/E = Switch Board | X = Exit"))
}

func (r *Renderer) DrawGameOver() {
	r.ClearScreen()
	r.DrawBoard()
	fmt.Println(r.CenterString("Game Over!"))
	fmt.Printf(r.CenterString("Final Score: %.0f\n"), r.game.score)
}
