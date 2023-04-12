package unbroken

import "time"

// GoTestLine represents a single line of output from a go test run
type GoTestLine struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Output  string    `json:"Output"`
	Elapsed float64   `json:"Elapsed"`
	Test    string    `json:"Test"`
}
