package fs

var (
	challengeAIFileContent = []byte(`Write to this file to challenge AI.

Examples:
echo 'level:5 days:1' > /mnt/chess/correspondence/challenge-ai
echo '' > /mnt/chess/correspondence/challenge-ai

Supported fields:
- level:   AI level (1-8) - default is 6
- color:   player color - default is "random"
- days:    days per move - default is 3
`)
	challengeRandomFileContent = []byte(`Write to this file to challenge a random player.	

Examples:
echo 'days:1 rated:true' > /mnt/chess/correspondence/challenge-random
echo '' > /mnt/chess/correspondence/challenge-random

Supported fields:
- days:    days per move - default is 3
- rated:   days per move - default is true
`)

	moveFileContent = []byte(`Write to this file to make a move. The data must be a UCI formatted move.	

Examples:
echo 'e2e4' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/move
echo 'e7e8q' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/move
echo 'e1g1' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/move
`)

	resignFileContent = []byte(`Write to this file to resign from the game.	

Examples:
echo '' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/resign
`)

	drawFileContent = []byte(`Write to this file to offer, accept, or reject a draw for the game. The data must be one of offer, accept, or reject.

Examples:
echo 'offer' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/draw
echo 'accept' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/draw
echo 'reject' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/draw
`)

	takebackFileContent = []byte(`Write to this file to offer, accept, or reject a takeback. The data must be one of offer, accept, or reject.

Examples:
echo 'offer' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/takeback
echo 'accept' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/takeback
echo 'reject' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/takeback
`)

	abortFileContent = []byte(`Write to this file to abort the game.	

Examples:
echo '' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/abort
`)

	claimVictoryFileContent = []byte(`Write to this file to claim victory.	

Examples:
echo '' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/claim-victory
`)

	claimDrawFileContent = []byte(`Write to this file to claim draw.	

Examples:
echo '' > /mnt/chess/correspondence/Me(1500).Opponent(1500).Qa7FJNk2/claim-draw
`)
)
