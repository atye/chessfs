package env

import (
	"os"
	"strconv"
	"strings"

	"github.com/atye/chessfs/internal/domain"
)

var (
	boardStyle                     domain.BoardStyle
	maxFinishedCorrespondenceGames int
	lightSquareHex                 string
	darkSquareHex                  string
	darkMode                       bool
)

func init() {
	loadEnvConfig()
}

func loadEnvConfig() {
	boardStyle = domain.DefaultStyle
	v := os.Getenv(domain.BoardStyleEnv)
	if v != "" {
		if domain.BoardStyle(v).Valid() {
			boardStyle = domain.BoardStyle(v)
		}
	}

	maxFinishedCorrespondenceGames = domain.DefaultMaxFinishedCorrespondenceGames
	v = os.Getenv(domain.MaxFinishedCorrespondenceGamesEnv)
	if v != "" {
		val, err := strconv.Atoi(v)
		if err != nil {
			return
		}

		if val < 1 {
			return
		}
		maxFinishedCorrespondenceGames = val
	}

	lightSquareHex = domain.DefaultLightSquareHex
	v = os.Getenv(domain.LightSquareColorEnv)
	if v != "" {
		lightSquareHex = v
		if !strings.HasPrefix(lightSquareHex, "#") {
			lightSquareHex = "#" + lightSquareHex
		}
	}

	darkSquareHex = domain.DefaultDarkSquareHex
	v = os.Getenv(domain.DarkSquareColorEnv)
	if v != "" {
		darkSquareHex = v
		if !strings.HasPrefix(darkSquareHex, "#") {
			darkSquareHex = "#" + darkSquareHex
		}
	}

	darkMode = domain.DefaultDarkMode
	v = os.Getenv(domain.DarkModeEnv)
	var err error
	if v != "" {
		darkMode, err = strconv.ParseBool(v)
		if err != nil {
			darkMode = domain.DefaultDarkMode
		}
	}
}

func GetBoardStyle() domain.BoardStyle {
	return boardStyle
}

func GetMaxFinishedCorrespondenceGames() int {
	return maxFinishedCorrespondenceGames
}

func GetLightSquareColor() string {
	return lightSquareHex
}

func GetDarkSquareColor() string {
	return darkSquareHex
}

func GetDarkMode() bool {
	return darkMode
}
