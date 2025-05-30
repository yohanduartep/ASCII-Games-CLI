package main

import (
	"math/rand"
	"time"
)

type Game struct {
	board             [HEIGHT][WIDTH]int
	boards            map[string][HEIGHT][WIDTH]int
	currentBoard      string
	currentPiece      [][]int
	fallInterval      float64
	gameTicker        *time.Ticker
	gameWidth         int
	level             int
	linesForNextLevel int
	nextPiece         [][]int
	piecePosition     [2]int
	rng               *rand.Rand
	score             float64
	terminalWidth     int
	totalLinesCleared int
}

func NewGame() *Game {
	game := &Game{
		boards:            make(map[string][HEIGHT][WIDTH]int),
		currentBoard:      "center",
		fallInterval:      10,
		gameWidth:         0,
		level:             0,
		linesForNextLevel: 10,
		score:             0,
		terminalWidth:     80,
		totalLinesCleared: 0,
	}

	game.boards["left"] = [HEIGHT][WIDTH]int{}
	game.boards["center"] = [HEIGHT][WIDTH]int{}
	game.boards["right"] = [HEIGHT][WIDTH]int{}
	game.board = game.boards[game.currentBoard]
	game.gameWidth = 3*(WIDTH*len(BLOCK_CELL)+2) + 2
	game.rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	return game
}

func (g *Game) StartTicker() *time.Ticker {
	g.gameTicker = time.NewTicker(time.Duration(float64(time.Second) * g.fallInterval / float64(FPS)))
	return g.gameTicker
}
