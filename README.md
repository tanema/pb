# pb The command line puzzle box
pb is the worlds first command line puzzle box. It is like a capture the flag
challenge but without the need for a network connection. There are several
stages to the puzzle that will challenge you in different ways

## Current state
There are currently 4 puzzles to make your way through

| Stage | Title |
|-------|-------|
|1      | Let's go to the movies. |
|2      | A Conversation |
|3      | Speech Impediment |
|4      | Merry-go-round |

## Usage
There are different usages for each stage so remember to check the help text for
each stage using the `--help` flag. Also each stage has different hints if you
get stuck so check those out with the `--hint` flag.

**full disclosure:**
This app will make artifacts around your system and `pb` tracks these completely
so that you can ensure that it is not doing anything funky and you can get rid
of them. At anytime you can run, `pb --artifacts` to see what they are so that
you can clean them up if you want to (or even check out what the contents are).

## Installation


```bash
> go install github.com/tanema/pb
```
