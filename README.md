# tuipardy

a terminal-based jeopardy board

## features

- classic jeopardy-style game board (sans daily-double)
- support for multiple teams
- score tracking and modification
- image support (with kitty, extensible to iterm2 and sixel terminals)

## question format

```csv
category,value,question,answer,imagepath
algorithms,200,"question","answer",image.png
```

image path is relative to the executable

## build

Prerequisites:

- Go 1.24+ (see `go.mod`)
- A Kitty-graphics–capable terminal for image rendering (e.g., Kitty, WezTerm with Kitty graphics). Non‑Kitty terminals will fall back to text-only questions.

Build the binary:

```bash
go build -o tuipardy .
```

## run

From the repository root, run the game by pointing it at a board CSV:

```bash
./tuipardy questions/board.csv
```

On macOS/Linux, the binary is `./tuipardy`; on Windows, use `tuipardy.exe`.

## prepare your board

- The CSV schema is:
  - `category`: string
  - `value`: integer (the dollar amount shown on the board)
  - `question`: string
  - `answer`: string
  - `imagepath` (optional): path to an image file for that question
- Paths in `imagepath` are resolved relative to where you run the binary. A simple convention is to place images in `questions/images/` and reference them like `questions/images/myimage.png`.
- The sample board lives at `questions/board.csv`. You can edit it directly or replace it with your own file.

Board size constraints (defaults):

- Categories: 6 (`ExpectedCategories` in `types.go`)
- Questions per category: 3 (`QuestionsPerCategory` in `types.go`)

If you want a different board size, change the constants in `types.go` and rebuild:

```go
// in types.go
const (
    QuestionsPerCategory = 3
    ExpectedCategories   = 6
)
```

Make sure your CSV matches those counts; the loader will error if they don’t.

## images

- Images are rendered only if your terminal supports Kitty graphics.
- Supported formats include PNG, JPEG, and GIF.
- Use the optional 5th CSV column (`imagepath`) to attach an image to a question, e.g.:
  ```csv
  "Algorithms",200,"What is Big-O of binary search?","O(log n)","questions/images/binary.png"
  ```

## controls

- Arrow keys or `h` `j` `k` `l` to move around the board
- `Enter` on the board: open the selected question
- `Space`/`Enter` on a question: toggle between question and answer
- `Esc`: go back to the board
- Score changes on the board: type `<teamNumber><+|-><value>` then press `Enter`
  - Example: `1+200` adds 200 to Team 1; `2-100` subtracts 100 from Team 2
- `q` or `Ctrl-C`: quit
