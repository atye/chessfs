package fs

var fileContent = map[string][]byte{
	"challenge-ai":     challengeAI,
	"challenge-random": challengeRandom,
	"move":             move,
	"resign":           resign,
	"draw":             draw,
	"takeback":         takeback,
	"abort":            abort,
	"claim-victory":    claimVictory,
	"claim-draw":       claimDraw,
}

var (
	challengeAI = []byte(`Write to this file to challenge AI.

Examples:
echo 'level:5 days:1' > /mnt/chess/correspondence/challenge-ai
echo '' > /mnt/chess/correspondence/challenge-ai

Supported fields:
- level:   AI level (1-8) - default is 6
- color:   player color - default is "random"
- days:    days per move - default is 3
`)
	challengeRandom = []byte(`Write to this file to challenge a random player.	

Examples:
echo 'days:1 rated:true' > /mnt/chess/correspondence/challenge-random
echo '' > /mnt/chess/correspondence/challenge-random

Supported fields:
- days:    days per move - default is 3
- rated:   days per move - default is true
`)

	move = []byte(`Write to this file to make a move. The data must be a UCI formatted move.	

Examples:
echo 'e2e4' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/move
echo 'e7e8q' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/move
echo 'e1g1' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/move
`)

	resign = []byte(`Write to this file to resign from the game.	

Examples:
echo '' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/resign
`)

	draw = []byte(`Write to this file to offer, accept, or reject a draw for the game. The data must be one of offer, accept, or reject.

Examples:
echo 'offer' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/draw
echo 'accept' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/draw
echo 'reject' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/draw
`)

	takeback = []byte(`Write to this file to offer, accept, or reject a takeback. The data must be one of offer, accept, or reject.

Examples:
echo 'offer' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/takeback
echo 'accept' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/takeback
echo 'reject' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/takeback
`)

	abort = []byte(`Write to this file to abort the game.	

Examples:
echo '' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/abort
`)

	claimVictory = []byte(`Write to this file to claim victory.	

Examples:
echo '' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/claim-victory
`)

	claimDraw = []byte(`Write to this file to claim draw.	

Examples:
echo '' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/claim-draw
`)
)
