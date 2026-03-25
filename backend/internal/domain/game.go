// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package domain

type Game struct {
	Races            []Race
	ActiveHomeWorldID PlanetID
	Empires          []Empire
}

type EmpireID int

type Empire struct {
	ID        EmpireID
	Name      string
	Active    bool
	Race      RaceID
	HomeWorld PlanetID
	Colonies  []ColonyID
	Ships     []ShipID
}

type RaceID int

type Race struct {
	ID        RaceID
	HomeWorld PlanetID
	Empires   []EmpireID
}
