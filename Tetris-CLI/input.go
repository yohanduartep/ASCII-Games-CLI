package main

import (
	"fmt"
	"golang.org/x/term"
	"os"
)

type InputHandler struct {
	game          *Game
	logic         *GameLogic
	rend          *Renderer
	userInputChan chan rune
}

func NewInputHandler(game *Game, logic *GameLogic, rend *Renderer) *InputHandler {
	return &InputHandler{
		game:          game,
		logic:         logic,
		rend:          rend,
		userInputChan: make(chan rune, 1),
	}
}

func (i *InputHandler) ListenForInput() {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		fmt.Fprintln(os.Stderr, "Input is not a terminal. Keyboard input disabled.")
		close(i.userInputChan)
		return
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set terminal to raw mode: %v\n", err)
		close(i.userInputChan)
		return
	}
	defer term.Restore(fd, oldState)

	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			close(i.userInputChan)
			return
		}
		i.userInputChan <- rune(buf[0])
	}
}

func (i *InputHandler) HandleInput(input rune) bool {
	switch input {
	case 'e', 'E':
		switch i.game.currentBoard {
		case "left":
			i.logic.SwitchBoard("center")
		case "center":
			i.logic.SwitchBoard("right")
		case "right":
			i.logic.SwitchBoard("left")
		}
	case 'q', 'Q':
		switch i.game.currentBoard {
		case "left":
			i.logic.SwitchBoard("right")
		case "right":
			i.logic.SwitchBoard("center")
		case "center":
			i.logic.SwitchBoard("left")
		}
	case 'a', 'A':
		i.logic.MoveHorizontally(-1)
	case 'd', 'D':
		i.logic.MoveHorizontally(1)
	case 's', 'S':
		if !i.logic.MoveDown() {
			i.logic.LockPiece()
			i.logic.SpawnPiece()
		}
	case 'w', 'W':
		i.logic.Rotate()
	case 'f', 'F':
		i.logic.HardDrop()
	case 'x', 'X':
		return false
	default:
		return true
	}
	return true
}
