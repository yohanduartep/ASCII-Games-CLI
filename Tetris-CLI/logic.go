package main

import (
	"os"
	"time"
)

type GameLogic struct {
	game *Game
	rend *Renderer
}

func NewGameLogic(game *Game, rend *Renderer) *GameLogic {
	return &GameLogic{
		game: game,
		rend: rend,
	}
}

func (l *GameLogic) ClearLines() {
	if l.game.currentBoard != "center" {
		return
	}
	tempScore := 0.0
	multiplier := 1.0
	linesCleared := 0

	for row := HEIGHT - 1; row >= 0; {
		isLineFull := true
		for c := range WIDTH {
			if l.game.board[row][c] == 0 {
				isLineFull = false
				break
			}
		}

		if isLineFull {
			linesCleared++
			for moveR := row; moveR > 0; moveR-- {
				copy(l.game.board[moveR][:], l.game.board[moveR-1][:])
			}
			l.game.board[0] = [WIDTH]int{}
		} else {
			row--
		}
	}

	if linesCleared > 0 {
		l.game.totalLinesCleared += linesCleared
		if l.game.totalLinesCleared >= l.game.linesForNextLevel {
			l.game.level++
			l.game.linesForNextLevel = (l.game.level + 1) * 10
			l.UpdateFallSpeed()
		}
		if linesCleared == 1 {
			tempScore = 100
		} else if linesCleared == 2 {
			tempScore = 300
		} else if linesCleared == 3 {
			tempScore = 500
		} else {
			tempScore = 850
		}
		l.game.boards[l.game.currentBoard] = l.game.board

		leftMultiplier := l.ScoreMultipliers("left")
		rightMultiplier := l.ScoreMultipliers("right")
		multiplier = (leftMultiplier * rightMultiplier)

		tempScore *= multiplier
		l.game.score += tempScore
		l.SwitchBoard("center")
	}

}

func (l *GameLogic) SpawnPiece() {
	l.game.currentPiece = l.game.nextPiece
	if len(l.game.currentPiece) > 0 && len(l.game.currentPiece[0]) > 0 {
		l.game.piecePosition = [2]int{0, WIDTH/2 - len(l.game.currentPiece[0])/2}
	} else {
		l.game.piecePosition = [2]int{0, WIDTH / 2}
		l.game.currentPiece = [][]int{{1}}
	}

	if !l.IsValidState(l.game.currentPiece, l.game.piecePosition) {
		l.rend.ClearScreen()
		l.rend.DrawBoard()
		os.Exit(0)
	}
	l.SetNext()
}

func (l *GameLogic) MoveDown() bool {
	nextPosition := [2]int{l.game.piecePosition[0] + 1, l.game.piecePosition[1]}
	if l.IsValidState(l.game.currentPiece, nextPosition) {
		l.game.piecePosition = nextPosition
		return true
	}
	return false
}

func (l *GameLogic) MoveHorizontally(deltaX int) {
	nextPosition := [2]int{l.game.piecePosition[0], l.game.piecePosition[1] + deltaX}
	if l.IsValidState(l.game.currentPiece, nextPosition) {
		l.game.piecePosition = nextPosition
	}
}

func (l *GameLogic) LockPiece() {
	if l.game.currentPiece == nil {
		return
	}
	for rowIdx, rowData := range l.game.currentPiece {
		for cellIdx, cell := range rowData {
			if cell == 1 {
				boardRow := l.game.piecePosition[0] + rowIdx
				boardCol := l.game.piecePosition[1] + cellIdx
				if boardRow >= 0 && boardRow < HEIGHT && boardCol >= 0 && boardCol < WIDTH {
					l.game.board[boardRow][boardCol] = 1
				}
			}
		}
	}
	l.ClearLines()
	l.game.boards[l.game.currentBoard] = l.game.board
}

func (l *GameLogic) HardDrop() {
	for l.MoveDown() {
	}
	l.LockPiece()
	l.SpawnPiece()
}

func (l *GameLogic) Rotate() {
	if l.game.currentPiece == nil || len(l.game.currentPiece) == 0 || len(l.game.currentPiece[0]) == 0 {
		return
	}

	originalRows := len(l.game.currentPiece)
	originalCols := len(l.game.currentPiece[0])
	rotatedPiece := make([][]int, originalCols)

	for i := range rotatedPiece {
		rotatedPiece[i] = make([]int, originalRows)
	}

	for r := range l.game.currentPiece {
		for c := range l.game.currentPiece[r] {
			rotatedPiece[c][originalRows-1-r] = l.game.currentPiece[r][c]
		}
	}

	if l.IsValidState(rotatedPiece, l.game.piecePosition) {
		l.game.currentPiece = rotatedPiece
	}

}

func (l *GameLogic) SwitchBoard(newBoard string) {
	l.game.boards[l.game.currentBoard] = l.game.board
	targetBoard := l.game.boards[newBoard]
	if l.isValidOnBoard(l.game.currentPiece, l.game.piecePosition, targetBoard) {
		l.game.currentBoard = newBoard
		l.game.board = l.game.boards[l.game.currentBoard]
	}
}

func (l *GameLogic) isValidOnBoard(piece [][]int, pos [2]int, board [HEIGHT][WIDTH]int) bool {
	for rowIdx, row := range piece {
		for colIdx, cell := range row {
			if cell == 1 {
				boardRow := pos[0] + rowIdx
				boardCol := pos[1] + colIdx

				if boardRow < 0 || boardRow >= HEIGHT || boardCol < 0 || boardCol >= WIDTH {
					return false
				}
				if board[boardRow][boardCol] == 1 {
					return false
				}
			}
		}
	}
	return true
}

func (l *GameLogic) UpdateFallSpeed() {
	baseFallInterval := INITIAL_FALL_INTERVAL
	newSpeed := max(0.0, baseFallInterval-(float64(l.game.level)*0.5))
	l.game.fallInterval = newSpeed

	if l.game.gameTicker != nil {
		l.game.gameTicker.Reset(time.Duration(float64(time.Second) * l.game.fallInterval / float64(FPS)))
	}
}

func (l *GameLogic) IsValidState(pieceToCheck [][]int, pos [2]int) bool {
	if pieceToCheck == nil {
		return false
	}
	for rowIdx, rowData := range pieceToCheck {
		for cellIdx, cell := range rowData {
			if cell == 1 {
				boardRow := pos[0] + rowIdx
				boardCol := pos[1] + cellIdx

				if boardRow < 0 || boardRow >= HEIGHT || boardCol < 0 || boardCol >= WIDTH {
					return false
				}
				if l.game.board[boardRow][boardCol] == 1 {
					return false
				}
			}
		}
	}
	return true
}

func (l *GameLogic) ScoreMultipliers(boardName string) float64 {
	tempCurrentBoard := l.game.currentBoard
	l.SwitchBoard(boardName)

	linesCleared := 0
	for r := HEIGHT - 1; r >= 0; {
		isLineFull := true
		for c := range WIDTH {
			if l.game.board[r][c] == 0 {
				isLineFull = false
				break
			}
		}

		if isLineFull {
			linesCleared++
			for moveR := r; moveR > 0; moveR-- {
				copy(l.game.board[moveR][:], l.game.board[moveR-1][:])
			}
			l.game.board[0] = [WIDTH]int{}
		} else {
			r--
		}
	}

	l.game.boards[boardName] = l.game.board
	l.SwitchBoard(tempCurrentBoard)

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

func (l *GameLogic) SetNext() {
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
	for shape := range shapes {
		keys = append(keys, shape)
	}

	if len(keys) > 0 {
		randomKey := keys[l.game.rng.Intn(len(keys))]
		l.game.nextPiece = shapes[randomKey]
	} else {
		l.game.nextPiece = [][]int{{1}}
	}
}
