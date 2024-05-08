package stringer

//go:generate stringer -type=Pill

//go:generate enumer -type=Pill -sql -json -output=pill_enumer.go -transform=snake
type Pill int

const (
	Placebo Pill = iota
	Aspirin
	Ibuprofen
	Paracetamol
)
