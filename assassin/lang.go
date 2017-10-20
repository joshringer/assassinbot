package assassin

import "fmt"

// Lang represents localised template strings for the game
type Lang struct {
	EIP,
	GS, GE, GW, GWM, GD,
	PA, PD, PT, PCS, PAS string
}

// Fmt should be used to format a template string when substitutions are required.
func (l Lang) Fmt(t string, a ...interface{}) string {
	return fmt.Sprintf(t, a...)
}

// LangEn : English template strings
var LangEn = Lang{
	EIP: "Game already in progress",
	GS:  "The game has begun.",
	GE:  "The game has ended.",
	GW:  "%v wins.",
	GWM: "Surviving this time: %v.",
	GD:  "%v has been assassinated.",
	PA:  "You are alive.",
	PD:  "You have been assassinated.",
	PT:  "Your target is %v. Your KillWord is %v.",
	PAS: "Your attack was successful.",
	PCS: "Your counterattack was successful.",
}
