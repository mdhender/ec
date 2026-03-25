// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package domain

type Game struct {
	Cluster Cluster
	Empires []Empire
}

type EmpireID int

type Empire struct {
	ID        EmpireID
	Name      string
	HomeWorld PlanetID
	Colonies  []ColonyID
	Ships     []ShipID
}
