// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package domain

import "errors"

// OrderKind identifies the type of an order.
type OrderKind int

const (
	OrderKindSetUp        OrderKind = iota + 1
	OrderKindBuildChange
	OrderKindMiningChange
	OrderKindTransfer
	OrderKindAssemble
	OrderKindMove
	OrderKindDraft
	OrderKindPay
	OrderKindRation
	OrderKindName
)

func (k OrderKind) Valid() bool {
	return k >= OrderKindSetUp && k <= OrderKindName
}

func (k OrderKind) String() string {
	switch k {
	case OrderKindSetUp:
		return "set-up"
	case OrderKindBuildChange:
		return "build-change"
	case OrderKindMiningChange:
		return "mining-change"
	case OrderKindTransfer:
		return "transfer"
	case OrderKindAssemble:
		return "assemble"
	case OrderKindMove:
		return "move"
	case OrderKindDraft:
		return "draft"
	case OrderKindPay:
		return "pay"
	case OrderKindRation:
		return "ration"
	case OrderKindName:
		return "name"
	default:
		return "unknown"
	}
}

// Phase is the turn phase number (1–21) in which an order executes.
type Phase int

func (p Phase) Valid() bool {
	switch p {
	case PhaseSetUp, PhaseBuildChange, PhaseMiningChange, PhaseTransfer,
		PhaseAssemble, PhaseMove, PhaseDraft, PhasePay, PhaseName:
		// PhaseRation is omitted because it equals PhasePay (both 16).
		return true
	default:
		return false
	}
}

const (
	PhaseSetUp        Phase = 4
	PhaseBuildChange  Phase = 6
	PhaseMiningChange Phase = 7
	PhaseTransfer     Phase = 8
	PhaseAssemble     Phase = 9
	PhaseMove         Phase = 14
	PhaseDraft        Phase = 15
	PhasePay          Phase = 16
	PhaseRation       Phase = 16
	PhaseName         Phase = 19
)

// NameTargetKind identifies what kind of entity a NameOrder renames.
type NameTargetKind int

const (
	NameTargetPlanet NameTargetKind = iota + 1
	NameTargetShip
	NameTargetColony
)

func (k NameTargetKind) Valid() bool {
	return k >= NameTargetPlanet && k <= NameTargetColony
}

func (k NameTargetKind) String() string {
	switch k {
	case NameTargetPlanet:
		return "planet"
	case NameTargetShip:
		return "ship"
	case NameTargetColony:
		return "colony"
	default:
		return "unknown"
	}
}

// Order is the interface all typed order structs satisfy.
type Order interface {
	Kind() OrderKind
	TurnPhase() Phase
	Validate() error
}

// MoveDestination holds the raw parsed destination for a move order.
// Full execution-time resolution requires the ship-location model from a later sprint.
type MoveDestination struct {
	Raw string // e.g. "01-02-03/4" or "01-02-03"
}

// SetUpOrder creates a new ship or colony.
type SetUpOrder struct {
	OrderKind OrderKind
	ColonyID  ColonyID // source colony funding the set-up
	NewName   string
	NewKind   UnitSpec
}

func (o SetUpOrder) Kind() OrderKind  { return o.OrderKind }
func (o SetUpOrder) TurnPhase() Phase { return PhaseSetUp }

func (o SetUpOrder) Validate() error {
	if o.OrderKind != OrderKindSetUp {
		return errors.New("set-up order: invalid order kind")
	}
	if o.ColonyID <= 0 {
		return errors.New("set-up order: colony ID must be positive")
	}
	if o.NewName == "" {
		return errors.New("set-up order: new name must not be empty")
	}
	if !o.NewKind.Kind.Valid() {
		return errors.New("set-up order: new unit kind must be valid")
	}
	return nil
}

// BuildChangeOrder redirects a factory group to produce a different unit kind.
type BuildChangeOrder struct {
	OrderKind      OrderKind
	ColonyID       ColonyID
	FactoryGroupID FactoryGroupID
	NewUnit        UnitSpec
}

func (o BuildChangeOrder) Kind() OrderKind  { return o.OrderKind }
func (o BuildChangeOrder) TurnPhase() Phase { return PhaseBuildChange }

func (o BuildChangeOrder) Validate() error {
	if o.OrderKind != OrderKindBuildChange {
		return errors.New("build-change order: invalid order kind")
	}
	if o.ColonyID <= 0 {
		return errors.New("build-change order: colony ID must be positive")
	}
	if o.FactoryGroupID <= 0 {
		return errors.New("build-change order: factory group ID must be positive")
	}
	if !o.NewUnit.Kind.Valid() {
		return errors.New("build-change order: new unit kind must be valid")
	}
	return nil
}

// MiningChangeOrder reassigns a mining group to a different deposit.
type MiningChangeOrder struct {
	OrderKind     OrderKind
	ColonyID      ColonyID
	MiningGroupID MiningGroupID
	DepositID     DepositID
}

func (o MiningChangeOrder) Kind() OrderKind  { return o.OrderKind }
func (o MiningChangeOrder) TurnPhase() Phase { return PhaseMiningChange }

func (o MiningChangeOrder) Validate() error {
	if o.OrderKind != OrderKindMiningChange {
		return errors.New("mining-change order: invalid order kind")
	}
	if o.ColonyID <= 0 {
		return errors.New("mining-change order: colony ID must be positive")
	}
	if o.MiningGroupID <= 0 {
		return errors.New("mining-change order: mining group ID must be positive")
	}
	if o.DepositID <= 0 {
		return errors.New("mining-change order: deposit ID must be positive")
	}
	return nil
}

// TransferOrder moves units between two colony locations.
type TransferOrder struct {
	OrderKind OrderKind
	SourceID  ColonyID
	DestID    ColonyID
	Unit      UnitSpec
	Quantity  int
}

func (o TransferOrder) Kind() OrderKind  { return o.OrderKind }
func (o TransferOrder) TurnPhase() Phase { return PhaseTransfer }

func (o TransferOrder) Validate() error {
	if o.OrderKind != OrderKindTransfer {
		return errors.New("transfer order: invalid order kind")
	}
	if o.SourceID <= 0 {
		return errors.New("transfer order: source ID must be positive")
	}
	if o.DestID <= 0 {
		return errors.New("transfer order: destination ID must be positive")
	}
	if !o.Unit.Kind.Valid() {
		return errors.New("transfer order: unit kind must be valid")
	}
	if o.Quantity <= 0 {
		return errors.New("transfer order: quantity must be positive")
	}
	return nil
}

// AssembleOrder assembles generic units into an operational group on a colony or ship.
// Syntax: assemble <id> <unit-token> <qty>
type AssembleOrder struct {
	OrderKind OrderKind
	ColonyID  ColonyID
	Unit      UnitSpec
	Quantity  int
}

func (o AssembleOrder) Kind() OrderKind  { return o.OrderKind }
func (o AssembleOrder) TurnPhase() Phase { return PhaseAssemble }

func (o AssembleOrder) Validate() error {
	if o.OrderKind != OrderKindAssemble {
		return errors.New("assemble order: invalid order kind")
	}
	if o.ColonyID <= 0 {
		return errors.New("assemble order: colony ID must be positive")
	}
	if !o.Unit.Kind.Valid() {
		return errors.New("assemble order: unit kind must be valid")
	}
	if o.Quantity <= 0 {
		return errors.New("assemble order: quantity must be positive")
	}
	return nil
}

// AssembleFactoryOrder assembles factory units configured to produce a specific unit kind.
// Syntax: assemble <id> factory <factory-unit> <qty> <build-target>
type AssembleFactoryOrder struct {
	OrderKind   OrderKind
	LocationID  ColonyID
	FactoryUnit UnitSpec // must be Factory
	FactoryQty  int
	BuildTarget UnitSpec
}

func (o AssembleFactoryOrder) Kind() OrderKind  { return o.OrderKind }
func (o AssembleFactoryOrder) TurnPhase() Phase { return PhaseAssemble }

func (o AssembleFactoryOrder) Validate() error {
	if o.OrderKind != OrderKindAssemble {
		return errors.New("assemble-factory order: invalid order kind")
	}
	if o.LocationID <= 0 {
		return errors.New("assemble-factory order: location ID must be positive")
	}
	if o.FactoryUnit.Kind != Factory {
		return errors.New("assemble-factory order: factory unit must be a factory")
	}
	if o.FactoryQty <= 0 {
		return errors.New("assemble-factory order: factory quantity must be positive")
	}
	if !o.BuildTarget.Kind.Valid() {
		return errors.New("assemble-factory order: build target must be valid")
	}
	return nil
}

// AssembleMineOrder assembles mine units assigned to a specific deposit.
// Syntax: assemble <id> mine <mine-unit> <qty> <deposit-id>
type AssembleMineOrder struct {
	OrderKind  OrderKind
	LocationID ColonyID
	MineUnit   UnitSpec // must be Mine
	MineQty    int
	DepositID  DepositID
}

func (o AssembleMineOrder) Kind() OrderKind  { return o.OrderKind }
func (o AssembleMineOrder) TurnPhase() Phase { return PhaseAssemble }

func (o AssembleMineOrder) Validate() error {
	if o.OrderKind != OrderKindAssemble {
		return errors.New("assemble-mine order: invalid order kind")
	}
	if o.LocationID <= 0 {
		return errors.New("assemble-mine order: location ID must be positive")
	}
	if o.MineUnit.Kind != Mine {
		return errors.New("assemble-mine order: mine unit must be a mine")
	}
	if o.MineQty <= 0 {
		return errors.New("assemble-mine order: mine quantity must be positive")
	}
	if o.DepositID <= 0 {
		return errors.New("assemble-mine order: deposit ID must be positive")
	}
	return nil
}

// MoveOrder moves a ship to a destination.
type MoveOrder struct {
	OrderKind   OrderKind
	ShipID      ShipID
	Destination MoveDestination
}

func (o MoveOrder) Kind() OrderKind  { return o.OrderKind }
func (o MoveOrder) TurnPhase() Phase { return PhaseMove }

func (o MoveOrder) Validate() error {
	if o.OrderKind != OrderKindMove {
		return errors.New("move order: invalid order kind")
	}
	if o.ShipID <= 0 {
		return errors.New("move order: ship ID must be positive")
	}
	if o.Destination.Raw == "" {
		return errors.New("move order: destination must not be empty")
	}
	return nil
}

// DraftOrder drafts population of a given kind into a colony's workforce.
type DraftOrder struct {
	OrderKind OrderKind
	ColonyID  ColonyID
	PopKind   UnitKind
	Quantity  int
}

func (o DraftOrder) Kind() OrderKind  { return o.OrderKind }
func (o DraftOrder) TurnPhase() Phase { return PhaseDraft }

func (o DraftOrder) Validate() error {
	if o.OrderKind != OrderKindDraft {
		return errors.New("draft order: invalid order kind")
	}
	if o.ColonyID <= 0 {
		return errors.New("draft order: colony ID must be positive")
	}
	if !o.PopKind.Valid() {
		return errors.New("draft order: population kind must be valid")
	}
	if o.Quantity <= 0 {
		return errors.New("draft order: quantity must be positive")
	}
	return nil
}

// PayOrder sets the wage rate for a population kind on a colony.
type PayOrder struct {
	OrderKind OrderKind
	ColonyID  ColonyID
	PopKind   UnitKind
	Wage      int // wage amount in game currency units; 0 is valid (no pay)
}

func (o PayOrder) Kind() OrderKind  { return o.OrderKind }
func (o PayOrder) TurnPhase() Phase { return PhasePay }

func (o PayOrder) Validate() error {
	if o.OrderKind != OrderKindPay {
		return errors.New("pay order: invalid order kind")
	}
	if o.ColonyID <= 0 {
		return errors.New("pay order: colony ID must be positive")
	}
	if !o.PopKind.Valid() {
		return errors.New("pay order: population kind must be valid")
	}
	if o.Wage < 0 {
		return errors.New("pay order: wage must not be negative")
	}
	return nil
}

// RationOrder sets the food ration percentage for a colony.
type RationOrder struct {
	OrderKind        OrderKind
	ColonyID         ColonyID
	RationPercentage int // 0–100
}

func (o RationOrder) Kind() OrderKind  { return o.OrderKind }
func (o RationOrder) TurnPhase() Phase { return PhaseRation }

func (o RationOrder) Validate() error {
	if o.OrderKind != OrderKindRation {
		return errors.New("ration order: invalid order kind")
	}
	if o.ColonyID <= 0 {
		return errors.New("ration order: colony ID must be positive")
	}
	if o.RationPercentage < 0 || o.RationPercentage > 100 {
		return errors.New("ration order: ration percentage must be between 0 and 100")
	}
	return nil
}

// NameOrder renames a planet, ship, or colony.
type NameOrder struct {
	OrderKind  OrderKind
	TargetKind NameTargetKind
	TargetID   int // PlanetID, ShipID, or ColonyID depending on TargetKind
	NewName    string
}

func (o NameOrder) Kind() OrderKind  { return o.OrderKind }
func (o NameOrder) TurnPhase() Phase { return PhaseName }

func (o NameOrder) Validate() error {
	if o.OrderKind != OrderKindName {
		return errors.New("name order: invalid order kind")
	}
	if !o.TargetKind.Valid() {
		return errors.New("name order: target kind must be valid")
	}
	if o.TargetID <= 0 {
		return errors.New("name order: target ID must be positive")
	}
	if o.NewName == "" {
		return errors.New("name order: new name must not be empty")
	}
	return nil
}
