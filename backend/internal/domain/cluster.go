// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package domain

import "math"

// Cluster contains all the objects in the cluster.
type Cluster struct {
	Systems  []System
	Stars    []Star
	Planets  []Planet
	Deposits []Deposit
	Colonies []Colony
	Ships    []Ship
}

type SystemID int

type System struct {
	ID       SystemID
	Display  string // Display label: "XX-YY-ZZ"
	Location Coords
	Stars    []StarID // ordered by sequence
}

type Coords struct {
	X int
	Y int
	Z int
}

// Distance returns the Euclidean distance between two coordinates.
func (c Coords) Distance(other Coords) float64 {
	dx := float64(c.X - other.X)
	dy := float64(c.Y - other.Y)
	dz := float64(c.Z - other.Z)
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

func (c Coords) Less(c2 Coords) bool {
	if c.X != c2.X {
		return c.X < c2.X
	}
	if c.Y != c2.Y {
		return c.Y < c2.Y
	}
	return c.Z < c2.Z
}

type StarID int

type Star struct {
	ID       StarID
	Sequence int
	Display  string       // Display label: "XX-YY-ZZ" or "XX-YY-ZZ/S" (per § 5).
	System   SystemID     // parent system
	Orbits   [10]PlanetID // index 0 = Orbit 1, index 9 = Orbit 10; value 0 means empty
}

type PlanetID int

type Planet struct {
	ID           PlanetID
	Kind         PlanetKind
	Habitability int
	Deposits     []DepositID
}

// PlanetKind classifies the contents of a star's orbit per § 4.1.
// Ordering matters: Terrestrial < AsteroidBelt < GasGiant is used when sorting.
type PlanetKind int

const (
	Terrestrial PlanetKind = iota + 1
	AsteroidBelt
	GasGiant
)

func (k PlanetKind) String() string {
	switch k {
	case Terrestrial:
		return "Terrestrial"
	case AsteroidBelt:
		return "Asteroid Belt"
	case GasGiant:
		return "Gas Giant"
	default:
		return "Unknown"
	}
}

type ColonyKind int

const (
	OpenAir ColonyKind = iota + 1
	Orbital
	Enclosed
)

func (k ColonyKind) String() string {
	switch k {
	case OpenAir:
		return "Open Air"
	case Orbital:
		return "Orbital"
	case Enclosed:
		return "Enclosed"
	default:
		return "Unknown"
	}
}

type DepositID int

type Deposit struct {
	ID                DepositID
	Resource          NaturalResource
	YieldPct          int // 1..100
	QuantityRemaining int
}

type NaturalResource int

const (
	GOLD NaturalResource = iota + 1
	FUEL
	METALLICS
	NONMETALLICS
)

func (r NaturalResource) String() string {
	switch r {
	case GOLD:
		return "Gold"
	case FUEL:
		return "Fuel"
	case METALLICS:
		return "Metallics"
	case NONMETALLICS:
		return "Non-Metallics"
	default:
		return "Unknown"
	}
}

type TechLevel int

// UnitSpec identifies a unit kind with an optional tech level.
// Units without tech level (food, population, etc.) have TechLevel 0.
type UnitSpec struct {
	Kind      UnitKind
	TechLevel TechLevel
}

type ColonyID int

type Colony struct {
	ID            ColonyID
	Name          string
	Empire        EmpireID
	Planet        PlanetID
	Kind          ColonyKind
	TechLevel     TechLevel
	Inventory     []Inventory
	MiningGroups  []MiningGroup
	FarmGroups    []FarmGroup
	FactoryGroups []FactoryGroup
}

type ShipID int

type Ship struct {
	ID        ShipID
	Name      string
	Empire    EmpireID
	Location  Coords
	TechLevel TechLevel
}

type UnitKind int

const (
	// Population
	Unemployables UnitKind = iota + 1
	UnskilledWorkers
	Professionals
	Soldiers
	Spies
	ConstructionWorkers
	Rebels

	// Weapons
	AssaultCraft
	AssaultWeapon
	AntiMissile
	EnergyShield
	EnergyWeapon
	MilitaryRobot
	MilitarySupply
	Missile
	MissileLauncher

	// Production
	Farm
	Factory
	Mine

	// Miscellaneous
	Automation
	ConsumerGoods
	Food
	HyperEngine
	LifeSupport
	LightStructural
	Sensor
	SpaceDrive
	Structural
	Transport

	// Uncategorized
	ResearchPoint
)

func (k UnitKind) Valid() bool {
	return k >= Unemployables && k <= ResearchPoint
}

type Inventory struct {
	Unit                 UnitKind
	TechLevel            TechLevel
	QuantityAssembled    int
	QuantityDisassembled int
}

// GroupUnit is a sub-group of a colony or ship group, representing
// all units of the same tech level assigned to that group.
type GroupUnit struct {
	TechLevel TechLevel
	Quantity  int
}

type MiningGroupID int

type MiningGroup struct {
	ID      MiningGroupID
	Deposit DepositID
	Units   []GroupUnit
}

type FarmGroupID int

// FarmGroup represents all farming units on a colony.
// Each colony has at most one FarmGroup; sub-groups are by tech level.
type FarmGroup struct {
	ID    FarmGroupID
	Units []GroupUnit
}

type FactoryGroupID int

type FactoryGroup struct {
	ID    FactoryGroupID
	Units []GroupUnit
}
