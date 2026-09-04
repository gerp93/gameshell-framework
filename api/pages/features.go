package apiPages

// AccountPageFeatures gates optional sections of the shared account page
// that not every game wants. Zero value is all-off, matching the safe
// default used by every other Set* parameterization point in this
// framework — a game that bumps its dependency without touching this call
// must not suddenly expose UI pointing at routes it never mounted.
type AccountPageFeatures struct {
	// WinCelebration shows the win-image/win-message upload section. Only
	// enable it if the game also mounts apiUser.SetWinGif/ClearWinGif/
	// GetWinGif/SetWinMessage.
	WinCelebration bool
	// LoseCelebration shows the lose-image/lose-message upload section, the
	// counterpart to WinCelebration. Only enable it if the game also mounts
	// apiUser.SetLoseGif/ClearLoseGif/GetLoseGif/SetLoseMessage.
	LoseCelebration bool
	// WinVideo shows the game-win YouTube clip section. Only enable it if
	// the game also mounts apiUser.SetWinVideo/ClearWinVideo.
	WinVideo bool
}

var accountPageFeatures AccountPageFeatures

// SetAccountPageFeatures overrides which optional account-page sections
// render. Call it once at startup, before serving requests.
func SetAccountPageFeatures(f AccountPageFeatures) {
	accountPageFeatures = f
}
