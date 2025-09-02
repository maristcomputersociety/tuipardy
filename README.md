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
