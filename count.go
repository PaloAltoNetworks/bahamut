// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import "fmt"

// Count holds various counter for a Context.
type Count struct {
	Total   int
	Current int
}

// newCount returns a new Count.
func newCount() *Count {

	return &Count{}
}

// Duplicate duplicates the Count.
func (c *Count) Duplicate() *Count {

	return &Count{
		Total:   c.Total,
		Current: c.Current,
	}
}

func (c *Count) String() string {

	return fmt.Sprintf("<count total:%d current:%d>",
		c.Total,
		c.Current,
	)
}
