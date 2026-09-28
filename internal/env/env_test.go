package env

import (
	"os"
	"testing"

	"github.com/atye/chessfs/internal/domain"
)

func setTestEnv(t *testing.T, values map[string]string) {
	t.Helper()
	for key, value := range values {
		oldValue, hadValue := os.LookupEnv(key)
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Setenv(%q, %q) failed: %v", key, value, err)
		}
		t.Cleanup(func() {
			if hadValue {
				_ = os.Setenv(key, oldValue)
				return
			}
			_ = os.Unsetenv(key)
		})
	}
}

func TestLoadEnvConfig_AllValuesSet(t *testing.T) {
	setTestEnv(t, map[string]string{
		"CHESSFS_BOARD_STYLE":                       "text",
		"CHESSFS_MAX_FINISHED_CORRESPONDENCE_GAMES": "7",
		"CHESSFS_LIGHT_SQUARE_HEX":                  "fff",
		"CHESSFS_DARK_SQUARE_HEX":                   "000",
		"CHESSFS_DARK_MODE":                         "false",
	})

	loadEnvConfig()

	if got, want := GetBoardStyle(), domain.Text; got != want {
		t.Fatalf("GetBoardStyle() = %q, want %q", got, want)
	}
	if got, want := GetMaxFinishedCorrespondenceGames(), 7; got != want {
		t.Fatalf("GetMaxFinishedCorrespondenceGames() = %d, want %d", got, want)
	}
	if got, want := GetLightSquareColor(), "#fff"; got != want {
		t.Fatalf("GetLightSquareColor() = %q, want %q", got, want)
	}
	if got, want := GetDarkSquareColor(), "#000"; got != want {
		t.Fatalf("GetDarkSquareColor() = %q, want %q", got, want)
	}
	if got, want := GetDarkMode(), false; got != want {
		t.Fatalf("GetDarkMode() = %v, want %v", got, want)
	}
}

func TestLoadEnvConfig_InvalidAndLowValuesUseDefaults(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  int
	}{
		{name: "non-numeric", value: "abc", want: domain.DefaultMaxFinishedCorrespondenceGames},
		{name: "zero", value: "0", want: domain.DefaultMaxFinishedCorrespondenceGames},
		{name: "negative", value: "-5", want: domain.DefaultMaxFinishedCorrespondenceGames},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setTestEnv(t, map[string]string{
				"CHESSFS_MAX_FINISHED_CORRESPONDENCE_GAMES": tc.value,
			})
			loadEnvConfig()
			if got := GetMaxFinishedCorrespondenceGames(); got != tc.want {
				t.Fatalf("GetMaxFinishedCorrespondenceGames() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestLoadEnvConfig_InvalidBoardStyleUsesDefault(t *testing.T) {
	setTestEnv(t, map[string]string{
		"CHESSFS_BOARD_STYLE": "nope",
	})
	loadEnvConfig()
	if got, want := GetBoardStyle(), domain.DefaultStyle; got != want {
		t.Fatalf("GetBoardStyle() = %q, want %q", got, want)
	}
}

func TestLoadEnvConfig_InvalidColorAndParseBoolUseDefaults(t *testing.T) {
	setTestEnv(t, map[string]string{
		"CHESSFS_LIGHT_SQUARE_HEX": "raw",
		"CHESSFS_DARK_SQUARE_HEX":  "raw",
		"CHESSFS_DARK_MODE":        "not-a-bool",
	})
	loadEnvConfig()
	if got, want := GetLightSquareColor(), "#raw"; got != want {
		t.Fatalf("GetLightSquareColor() = %q, want %q", got, want)
	}
	if got, want := GetDarkSquareColor(), "#raw"; got != want {
		t.Fatalf("GetDarkSquareColor() = %q, want %q", got, want)
	}
	if got, want := GetDarkMode(), domain.DefaultDarkMode; got != want {
		t.Fatalf("GetDarkMode() = %v, want %v", got, want)
	}
}
