package main

func main() {
	game := NewGame()
	renderer := NewRenderer(game)
	logic := NewGameLogic(game, renderer)
	logic.SetNext()
	input := NewInputHandler(game, logic, renderer)

	go input.ListenForInput()

	renderer.DrawBoard()

	ticker := game.StartTicker()
	defer ticker.Stop()

	for {
		select {
		case inputChar, ok := <-input.userInputChan:
			if !ok || !input.HandleInput(inputChar) {
				renderer.DrawGameOver()
				return
			}
			renderer.DrawBoard()
		case <-ticker.C:
			if !logic.MoveDown() {
				logic.LockPiece()
				logic.SpawnPiece()
			}
			renderer.DrawBoard()
		}
	}
}
