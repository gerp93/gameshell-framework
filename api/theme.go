package api

type Theme struct {
	Value string
	Label string
}

type ThemeGroup struct {
	Name   string
	Themes []Theme
}

var ThemeGroups = []ThemeGroup{
	{Name: "Classic", Themes: []Theme{
		{"dark-theme", "🌙 Dark"},
		{"light-theme", "☀️ Light"},
		{"nord-polar-night-theme", "❄️ Nord Polar Night"},
		{"dracula-theme", "🧛 Dracula"},
		{"purple-theme", "💜 Purple"},
		{"tokyo-night-dark-theme", "🌃 Tokyo Night Dark"},
		{"gruvbox-dark-theme", "🔳 GruvBox Dark"},
		{"gruvbox-light-theme", "🔲 GruvBox Light"},
	}},
	{Name: "Visual Assault", Themes: []Theme{
		{"retrowave-theme", "🌅 Retrowave"},
		{"bubblegum-theme", "🍬 Bubblegum"},
		{"electric-lime-theme", "⚡ Electric Lime"},
		{"neon-theme", "🌈 NEON"},
		{"commander-keen-theme", "🚀 Commander Keen"},
		{"lava-theme", "🌋 LAVA"},
		{"hacker-theme", "💻 Hacker"},
		{"hawkeye-theme", "🦅 Hawkeye"},
		{"merica-theme", "★ 'MERICA"},
	}},
	{Name: "Tractor", Themes: []Theme{
		{"green-acres-theme", "🚜 Green Acres"},
		{"red-barn-theme", "🏚️ Red Barn"},
		{"flambeau-theme", "🌾 Flambeau"},
		{"flambeau-inverse-theme", "🔥 Flambeau (Inverse)"},
		{"blue-oval-theme", "🔵 Blue Oval"},
	}},
}
