package util

func CheckSpelling(s *string) {
	spellchecked := *s + "[spelling checked]"
	*s = spellchecked
}
