package main

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
)

// game state
// - evens are whites
// - odds are blacks
// - holes are 0-11
// - mancalas are 12, and 13 respectively
// - state is on 14
type game [15]int

var opposite = map[int]int{
	0:  11,
	2:  9,
	4:  7,
	6:  5,
	8:  3,
	10: 1,
	11: 0,
	9:  2,
	7:  4,
	5:  6,
	3:  8,
	1:  10,
}

func (g game) String() string {
	var b bytes.Buffer
	white := g.isWhiteToPlay()
	var s string

	if white {
		s = " "
	} else {
		s = ">"
	}
	b.WriteString(fmt.Sprintf(
		"%s  | %2d | %2d | %2d | %2d | %2d | %2d |\n",
		s, g[11], g[9], g[7], g[5], g[3], g[1]))

	b.WriteString(fmt.Sprintf(
		"%2d +----+----+----+----+----+----+ %2d\n",
		g[13], g[12]))

	if !white {
		s = " "
	} else {
		s = ">"
	}
	b.WriteString(fmt.Sprintf(
		"%s  | %2d | %2d | %2d | %2d | %2d | %2d |\n",
		s, g[0], g[2], g[4], g[6], g[8], g[10]))
	return b.String()
}

func (g game) isWhiteToPlay() bool {
	return (g[14] & 0x1) == 0x0
}

func (g game) sum() int {
	sum := 0
	for i := 0; i < 14; i++ {
		sum += g[i]
	}
	return sum
}

func (g game) play(hole int) (game, error) {
	if g[hole] == 0 {
		return g, fmt.Errorf("no balls in this hole")
	}
	if (hole%2 == 0) != g.isWhiteToPlay() {
		return g, fmt.Errorf("out of sequence move")
	}

	var (
		side      = hole % 2
		remaining = g[hole]
		mancala   = (hole % 2) + 12
	)
	// fmt.Printf("mancala=%d\n", mancala)

	g[hole] = 0
	for 0 < remaining {
		hole += 2
		if mancala < hole {
			// 12 -> 1
			// 13 -> 0
			hole = (hole + 1) % 2
		}
		if hole < 12 || hole == mancala {
			g[hole]++
			remaining--
		}
	}

	// play again if you finish in the mancala
	if hole == mancala {
		g[14] = side
	} else {
		g[14] = (side + 1) % 2
	}

	// clear opposite hole if ends in previously empty hole on own side
	if hole != mancala && hole%2 == side && g[hole] == 1 {
		g[hole] = 0
		holeOp := opposite[hole]
		g[mancala] += 1 + g[holeOp]
		g[holeOp] = 0
	}

	return g, nil
}

func (g game) moves() []int {
	hole := 0
	if !g.isWhiteToPlay() {
		hole += 1
	}
	var moves []int
	for hole < 12 {
		if g[hole] != 0 {
			moves = append(moves, hole)
		}
		hole += 2
	}
	return moves
}

func (g game) finalscore() (int, int) {
	var white, black int
	for hole := 0; hole < 14; hole += 2 {
		white += g[hole]
	}
	for hole := 1; hole < 14; hole += 2 {
		black += g[hole]
	}
	return white, black
}

func newGame() game {
	var g game
	for i := 0; i < 12; i++ {
		g[i] = 4
	}
	return g
}

func (g game) score() int {
	var whites, blacks int
	for i := 0; i < 12; i += 2 {
		whites += g[i]
	}
	for i := 1; i < 12; i += 2 {
		blacks += g[i]
	}
	whiteBoost := 0
	if whites > blacks {
		whiteBoost = 2
	}
	blackBoost := 0
	if blacks > whites {
		blackBoost = 2
	}
	if g.isWhiteToPlay() {
		return g[12] + whiteBoost
	} else {
		return g[13] + blackBoost
	}
}

// minimax returns the optimal move index, and score
func (g game) minimax(depth int, white bool) (int, int) {
	var (
		moves     = g.moves()
		bestScore = math.MinInt32
		bestPlay  = 0
		max       = true
	)

	if depth == 0 || len(moves) == 0 {
		return -1, g.score()
	}

	//  isWhiteToPlay and  white then max
	// !isWhiteToPlay and  white then min
	//  isWhiteToPlay and !white then min
	// !isWhiteToPlay and !white then max

	if g.isWhiteToPlay() != white {
		max = false
	}
	if !max {
		bestScore = math.MaxInt32
	}

	for _, play := range moves {
		newG, err := g.play(play)
		if err != nil {
			panic(err)
		}
		_, score := newG.minimax(depth-1, white)
		if (max && bestScore < score) || (!max && score < bestScore) {
			bestScore = score
			bestPlay = play
		}
	}

	return bestPlay, bestScore
}

func askformove() int {
	for {
		fmt.Printf("your move (1 to 6): ")
		input := ""
		for {
			buf := make([]byte, 1)
			if _, err := io.ReadAtLeast(os.Stdin, buf, 1); err != nil {
				panic(err)
			}
			if string(buf[0]) == "\n" {
				break
			}
			input = input + string(buf[0])
		}
		move := input[0]
		if !('1' <= move && move <= '6') {
			fmt.Printf("... you must enter a number between 1 and 6\n")
			continue
		}
		return int(move-'1') + 1
	}
}

func main() {
	var (
		g   = newGame()
		err error
	)
	fmt.Printf("%s\n", g)

gameplay:
	for {
		for g.isWhiteToPlay() {
			if len(g.moves()) == 0 {
				break gameplay
			}
			move := askformove()

			hole := (move - 1) * 2
			g, err = g.play(hole)
			if err != nil {
				fmt.Printf("error: %s\n", err)
			}

			fmt.Printf("\n%s\n", g)
		}
		for !g.isWhiteToPlay() {
			if len(g.moves()) == 0 {
				break gameplay
			}
			fmt.Printf("now, it's my turn to play... let me think...\n")
			play, _ := g.minimax(11, false)
			g, err = g.play(play)
			if err != nil {
				panic(err)
			}
			fmt.Printf("\n%s\n", g)
		}
	}

	white, black := g.finalscore()
	if white < black {
		fmt.Printf("sorry, you lost %d to %d\n", white, black)
	} else if black < white {
		fmt.Printf("nice! you won %d to %d\n", white, black)
	} else {
		fmt.Printf("it's a draw")
	}
}
