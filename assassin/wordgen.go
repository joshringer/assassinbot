package assassin

/*
WordGenerator interface for generating player KillWords.
	Gen() generates and returns a single word
*/
type WordGenerator interface {
	Gen() string
}
