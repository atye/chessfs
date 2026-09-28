package chess

import (
	"errors"
	"fmt"
	"strings"

	"github.com/atye/chessfs/internal/domain"
	"github.com/atye/chessfs/internal/env"
	"github.com/charmbracelet/lipgloss"
	chess "github.com/corentings/chess/v3"
)

const (
	squareWidth  = 2
	squareHeight = 1
)

var (
	lightSquare = lipgloss.NewStyle().
			Background(lipgloss.Color(env.GetLightSquareColor()))

	darkSquare = lipgloss.NewStyle().
			Background(lipgloss.Color(env.GetDarkSquareColor()))

	whitePiece = lipgloss.Color("#FFFFFF")
	blackPiece = lipgloss.Color("#111111")
)

func Draw(uciMoves string, whitePerspective bool) (string, error) {
	switch env.GetBoardStyle() {
	case domain.Styled:
		return drawStyledBoard(uciMoves, whitePerspective)
	case domain.Text:
		return drawTextBoard(uciMoves, whitePerspective)
	default:
		return "", fmt.Errorf("invalid style")
	}
}

func drawTextBoard(uciMoves string, whitePerspective bool) (string, error) {
	game := chess.NewGame()

	for move := range strings.FieldsSeq(uciMoves) {
		if _, err := game.MoveText(move, chess.UCI(), nil); err != nil {
			return "", err
		}
	}

	if whitePerspective {
		return game.Position().Board().Draw2(chess.White, env.GetDarkMode()), nil
	}
	return game.Position().Board().Draw2(chess.Black, env.GetDarkMode()), nil
}

func drawStyledBoard(uciMoves string, whitePerspective bool) (string, error) {
	game := chess.NewGame()

	for move := range strings.FieldsSeq(uciMoves) {
		if _, err := game.MoveText(move, chess.UCI(), nil); err != nil {
			return "", err
		}
	}

	fields := strings.Fields(game.FEN())
	if len(fields) == 0 {
		return "", errors.New("no fields in FEN")
	}

	ranks := strings.Split(fields[0], "/")
	if len(ranks) != 8 {
		return "", errors.New("FEN does not contain 8 ranks")
	}

	var pieces [8][8]string

	for rank := 0; rank < 8; rank++ {
		file := 0

		for _, r := range ranks[rank] {
			if r >= '1' && r <= '8' {
				file += int(r - '0')
				continue
			}

			if file >= 8 {
				return "", fmt.Errorf(
					"too many files in rank %d",
					8-rank,
				)
			}

			pieces[rank][file] = string(r)
			file++
		}

		if file != 8 {
			return "", fmt.Errorf(
				"rank %d does not contain 8 files",
				8-rank,
			)
		}
	}

	var sb strings.Builder

	if whitePerspective {
		for rank := 0; rank < 8; rank++ {
			renderRank(&sb, pieces, rank, true)
		}
	} else {
		for rank := 7; rank >= 0; rank-- {
			renderRank(&sb, pieces, rank, false)
		}
	}

	renderFileLabels(&sb, whitePerspective)

	return sb.String(), nil
}

func renderRank(
	sb *strings.Builder,
	pieces [8][8]string,
	rank int,
	whitePerspective bool,
) {
	rankNumber := 8 - rank

	pieceRow := 0

	for row := 0; row < squareHeight; row++ {
		var line strings.Builder

		if row == pieceRow {
			fmt.Fprintf(&line, "%d ", rankNumber)
		} else {
			line.WriteString("  ")
		}

		for i := 0; i < 8; i++ {
			file := i

			if !whitePerspective {
				file = 7 - i
			}

			squareLines := strings.Split(
				renderSquare(pieces[rank][file], file, rank),
				"\n",
			)

			line.WriteString(squareLines[row])
		}

		sb.WriteString(line.String())
		sb.WriteByte('\n')
	}
}

func renderSquare(piece string, file, rank int) string {
	style := squareStyle(file, rank)

	if piece == "" {
		piece = " "
	}

	isWhite := piece >= "A" && piece <= "Z"
	piece = unicodePiece(piece)

	if piece != " " {
		if isWhite {
			style = style.Foreground(whitePiece)
		} else {
			style = style.Foreground(blackPiece)
		}
	}

	return style.Render(centerText(piece, squareWidth))
}

func squareStyle(file, rank int) lipgloss.Style {
	chessRank := 7 - rank

	if (file+chessRank)%2 == 0 {
		return darkSquare
	}

	return lightSquare
}

func renderFileLabels(sb *strings.Builder, whitePerspective bool) {
	sb.WriteString("  ")

	if whitePerspective {
		for file := 0; file < 8; file++ {
			sb.WriteString(centerText(
				string(rune('A'+file)),
				squareWidth,
			))
		}
	} else {
		for file := 7; file >= 0; file-- {
			sb.WriteString(centerText(
				string(rune('A'+file)),
				squareWidth,
			))
		}
	}

	sb.WriteByte('\n')
}

func centerText(s string, width int) string {
	padding := width - lipgloss.Width(s)

	if padding <= 0 {
		return s
	}

	left := padding / 2
	right := padding - left

	return strings.Repeat(" ", left) +
		s +
		strings.Repeat(" ", right)
}

func unicodePiece(piece string) string {
	switch piece {
	case "K", "k":
		return "♔"
	case "Q", "q":
		return "♕"
	case "R", "r":
		return "♖"
	case "B", "b":
		return "♗"
	case "N", "n":
		return "♘"
	case "P", "p":
		return "♙"

	default:
		return piece
	}
}
