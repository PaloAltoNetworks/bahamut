// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

// Count holds various counter for a context.
type Count struct {
	Total   int
	Current int
}

// NewCount creates a new Count.
func NewCount() *Count {

	return &Count{}
}
