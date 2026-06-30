package ui

import "time"

// TimeOfDay classifies the current hour into a named period.
type TimeOfDay int

const (
	Morning   TimeOfDay = iota // 05:00–11:59
	Afternoon                  // 12:00–16:59
	Evening                    // 17:00–20:59
	Night                      // 21:00–04:59
)

// CurrentTimeOfDay returns the TimeOfDay for now.
func CurrentTimeOfDay() TimeOfDay {
	h := time.Now().Hour()
	switch {
	case h >= 5 && h < 12:
		return Morning
	case h >= 12 && h < 17:
		return Afternoon
	case h >= 17 && h < 21:
		return Evening
	default:
		return Night
	}
}

// Scene returns the ASCII art scene for the given time of day.
func Scene(t TimeOfDay) string {
	switch t {
	case Morning:
		return morningScene
	case Afternoon:
		return afternoonScene
	case Evening:
		return eveningScene
	default:
		return nightScene
	}
}

// SceneLabel returns a label string for the time of day.
func SceneLabel(t TimeOfDay) string {
	switch t {
	case Morning:
		return "🌅 Morning"
	case Afternoon:
		return "☀️  Afternoon"
	case Evening:
		return "🌆 Evening"
	default:
		return "🌙 Night"
	}
}

const morningScene = `
    \\   /
     \\ /
   ---☀️---
     / \\
    /   \\

  ~~~~~ 🏡 ~~~~~
  [  working  ]
`

const afternoonScene = `
        ☀️
   . ☁️     ☁️ .
  .             .
  . .  🌳🌳🌳 . .
  [  deep work  ]
`

const eveningScene = `
  🌆       🌇
  |\  /\  /|
  | \/  \/ |
  |________|
  [ winding down ]
`

const nightScene = `
  ✦  ★    ✦   ★
    ★   ✦    ★
  ☽     🌌     ☽
  ★  ✦    ★  ✦
  [  burning midnight oil  ]
`
