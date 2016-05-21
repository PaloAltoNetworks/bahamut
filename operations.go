// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

// Operation represents a Cid operation.
type Operation int

const (
	// OperationRetrieveMany is the operation used to get multiple objects
	OperationRetrieveMany Operation = iota + 1

	// OperationRetrieve is the operation used to get a single object
	OperationRetrieve

	// OperationCreate is the operation used to create a single object
	OperationCreate

	// OperationUpdate is the operation used to update a single object
	OperationUpdate

	// OperationDelete is the operation used to delete a single object
	OperationDelete

	// OperationPatch is the operation used to patcj a single object
	OperationPatch

	// OperationInfo is the operation used to get info for a single object
	OperationInfo
)
