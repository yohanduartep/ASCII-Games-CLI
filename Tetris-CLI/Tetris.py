# Tetris Game in Python - ASCII Version. By Yohan
# This is an ASCII Tetris game implemented in Python.
# It uses the terminal for rendering and keyboard input.
# The game features three boards (left, center, right) and allows switching between them.

import os
import random
import select
import shutil
import sys
import termios
import time
import tty

WIDTH, HEIGHT = 10, 24
FALL_SPEED = 0.5

BOARD_LEFT = [[" " for _ in range(WIDTH)] for _ in range(HEIGHT)]
BOARD_CENTER = [[" " for _ in range(WIDTH)] for _ in range(HEIGHT)]
BOARD_RIGHT = [[" " for _ in range(WIDTH)] for _ in range(HEIGHT)]
BOARDS = {"left": BOARD_LEFT, "center": BOARD_CENTER, "right": BOARD_RIGHT}

SHAPES = {
    "I": [[1, 1, 1, 1]],
    "O": [[1, 1], [1, 1]],
    "T": [[0, 1, 0], [1, 1, 1]],
    "S": [[0, 1, 1], [1, 1, 0]],
    "Z": [[1, 1, 0], [0, 1, 1]],
    "J": [[1, 0, 0], [1, 1, 1]],
    "L": [[0, 0, 1], [1, 1, 1]],
}

current_board = "center"
points = 0
fall_speed = FALL_SPEED


def rotate(shape):
    return [
        [shape[row][col] for row in range(len(shape) - 1, -1, -1)]
        for col in range(len(shape[0]))
    ]


def clear_screen():
    os.system("clear" if os.name == "posix" else "cls")


def draw_shape(shape, x, y, board):
    temp_board = [row.copy() for row in board]
    for i, row in enumerate(shape):
        for j, cell in enumerate(row):
            if cell:
                bx, by = x + j, y + i
                if 0 <= by < HEIGHT and 0 <= bx < WIDTH:
                    temp_board[by][bx] = "#"
    return temp_board


def render_board(board):
    return "\n".join(
        "|" + "".join("[]" if cell == "#" else "  " for cell in row) + "|"
        for row in board
    )


def display_game(shape=None, x=0, y=0, next_shape=None):
    global points
    temp = (
        draw_shape(shape, x, y, BOARDS[current_board])
        if shape
        else BOARDS[current_board]
    )

    left = render_board(temp if current_board == "left" else BOARD_LEFT)
    center = render_board(temp if current_board == "center" else BOARD_CENTER)
    right = render_board(temp if current_board == "right" else BOARD_RIGHT)

    term_width = shutil.get_terminal_size((80, 24))[0]
    offset = max((term_width - (WIDTH * 6 + 6)) // 2, 0)

    combined = "\n".join(
        " " * offset + l + "   " + c + "   " + r
        for l, c, r in zip(left.split("\n"), center.split("\n"), right.split("\n"))
    )

    clear_screen()
    print("\n" * 2)
    indicator = {
        "left": (WIDTH) * " " + "\\/",
        "center": (WIDTH * 3) * " " + "     \\/",
        "right": (WIDTH * 5) * " " + "          \\/",
    }
    print(" " * offset + indicator[current_board])

    print(" " * offset + ("+" + "--" * WIDTH + "+   ") * 3)
    print(combined)
    print(" " * offset + ("+" + "--" * WIDTH + "+   ") * 3)

    if next_shape:
        preview = [[" " for _ in range(4)] for _ in range(4)]
        for i, row in enumerate(next_shape):
            for j, cell in enumerate(row):
                if cell:
                    preview[i][j] = "#"
        preview_offset = offset + WIDTH * 3
        print(" " * preview_offset + f"Points: {points}")
        print(" " * preview_offset + "+" + "--" * 4 + "+")
        print(" " * preview_offset + "|Next Set|")
        print(" " * preview_offset + "+" + "--" * 4 + "+")
        for row in preview:
            print(
                " " * preview_offset
                + "|"
                + "".join("[]" if c == "#" else "  " for c in row)
                + "|"
            )
        print(" " * preview_offset + "+" + "--" * 4 + "+")

    print(
        "\n"
        + " " * (offset)
        + "Controls: [a] left  [d] right  [s] drop  [w/r] rotate  [q/e] switch  [x] quit"
    )


def can_move(shape, x, y, board):
    for i, row in enumerate(shape):
        for j, cell in enumerate(row):
            if cell:
                bx, by = x + j, y + i
                if not (0 <= bx < WIDTH and 0 <= by < HEIGHT):
                    return False
                if board[by][bx] != " ":
                    return False
    return True


def place_shape(shape, x, y, board):
    for i, row in enumerate(shape):
        for j, cell in enumerate(row):
            if cell:
                bx, by = x + j, y + i
                if 0 <= bx < WIDTH and 0 <= by < HEIGHT:
                    board[by][bx] = "#"


def clear_lines(board):
    global points, fall_speed
    new_board = [row for row in board if " " in row]
    lines_cleared = HEIGHT - len(new_board)
    while len(new_board) < HEIGHT:
        new_board.insert(0, [" "] * WIDTH)
    board[:] = new_board

    if lines_cleared:
        points += [0, 100, 300, 500, 800][min(lines_cleared, 4)]
        if points // 1000 > (points - lines_cleared * 100) // 1000:
            fall_speed = max(0.01, fall_speed * 0.9)


def random_board_switch():
    global current_board
    current_board = random.choice(["left", "center", "right"])


def maybe_add_obstacles(last_time, interval):
    if time.time() - last_time > interval:
        random_board_switch()
        return time.time()
    return last_time


def read_key(timeout=0.1):
    fd = sys.stdin.fileno()
    old = termios.tcgetattr(fd)
    try:
        tty.setcbreak(fd)
        r, _, _ = select.select([fd], [], [], timeout)
        return sys.stdin.read(1) if r else None
    finally:
        termios.tcsetattr(fd, termios.TCSADRAIN, old)


def switch_board(target, shape, x, y):
    global current_board
    if can_move(shape, x, y, BOARDS[target]):
        current_board = target
    else:
        print("Game Over!")
        sys.exit()


def main_game():
    global current_board, fall_speed
    current_board = "center"
    next_shape = random.choice(list(SHAPES.values()))

    last_obstacle_time = time.time()
    obstacle_interval = 10

    while True:
        shape = next_shape
        next_shape = random.choice(list(SHAPES.values()))
        x = WIDTH // 2 - len(shape[0]) // 2
        y = 0

        if not can_move(shape, x, y, BOARDS[current_board]):
            print("Game Over!")
            break

        fall_timer = time.time()
        while True:
            display_game(shape, x, y, next_shape)
            key = read_key(0.05)

            obstacle_interval = max(2, 10 - points // 1000)
            last_obstacle_time = maybe_add_obstacles(
                last_obstacle_time, obstacle_interval
            )

            if key == "a" and can_move(shape, x - 1, y, BOARDS[current_board]):
                x -= 1
            elif key == "d" and can_move(shape, x + 1, y, BOARDS[current_board]):
                x += 1
            elif key == "s":
                while can_move(shape, x, y + 1, BOARDS[current_board]):
                    y += 1
                place_shape(shape, x, y, BOARDS[current_board])
                clear_lines(BOARDS[current_board])
                break
            elif key == "w" or key == "r":
                r = rotate(shape)
                if can_move(r, x, y, BOARDS[current_board]):
                    shape = r
            elif key == "q":
                switch_board(
                    {"left": "right", "center": "left", "right": "center"}[
                        current_board
                    ],
                    shape,
                    x,
                    y,
                )
            elif key == "e":
                switch_board(
                    {"left": "center", "center": "right", "right": "left"}[
                        current_board
                    ],
                    shape,
                    x,
                    y,
                )
            elif key == "x":
                return

            if time.time() - fall_timer > fall_speed:
                if can_move(shape, x, y + 1, BOARDS[current_board]):
                    y += 1
                else:
                    place_shape(shape, x, y, BOARDS[current_board])
                    clear_lines(BOARDS[current_board])
                    break
                fall_timer = time.time()


if __name__ == "__main__":
    main_game()
