// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package domain

// DepositTemplate describes a single natural resource deposit to create
// on a homeworld planet.
type DepositTemplate struct {
	Resource          NaturalResource
	YieldPct          int
	QuantityRemaining int
}

// HomeworldTemplate defines the starting conditions applied to a planet
// when it is designated as a homeworld. All homeworlds start with the
// same deposits and habitability.
type HomeworldTemplate struct {
	Habitability int
	Deposits     []DepositTemplate
}

// ColonyTemplate defines the starting state of the colony created when
// an empire is assigned to a homeworld.
type ColonyTemplate struct {
	Kind      ColonyKind
	TechLevel TechLevel
	Inventory []Inventory
}
