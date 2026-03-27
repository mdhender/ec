// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package ordertext implements a line-oriented text parser for the v0 order language.
package ordertext

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mdhender/ec/internal/app"
	"github.com/mdhender/ec/internal/domain"
)

// Parser implements app.OrderParser using a simple line-oriented tokenizer.
type Parser struct{}

// NewParser returns a new Parser.
func NewParser() *Parser { return &Parser{} }

// Parse parses raw order text line by line and returns typed domain orders and diagnostics.
// Parsing continues after bad lines so the full diagnostics set is returned.
// Keywords are case-insensitive. Line numbers in diagnostics are 1-based.
func (p *Parser) Parse(text string) ([]domain.Order, []app.ParseDiagnostic, error) {
	var orders []domain.Order
	var diags []app.ParseDiagnostic

	lines := strings.Split(text, "\n")
	for lineNo, raw := range lines {
		lineNum := lineNo + 1

		// strip comment: # starts a comment (after optional whitespace)
		if idx := strings.IndexByte(raw, '#'); idx >= 0 {
			raw = raw[:idx]
		}

		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		tokens := strings.Fields(line)

		order, diag, err := parseLine(lineNum, tokens)
		if err != nil {
			return nil, nil, err
		}
		if diag != nil {
			diags = append(diags, *diag)
		} else if order != nil {
			orders = append(orders, order)
		}
	}

	return orders, diags, nil
}

// parseLine dispatches a non-empty token slice to the appropriate order parser.
// Returns either an order, a diagnostic, or an error (for unexpected failures).
// Grammar uses keyword-first syntax: <command> <args...>
func parseLine(lineNum int, tokens []string) (domain.Order, *app.ParseDiagnostic, error) {
	keyword := strings.ToLower(tokens[0])
	rest := tokens[1:]

	switch keyword {
	case "setup":
		return nil, notImplemented(lineNum, "setup"), nil
	case "build":
		return parseBuildChange(lineNum, rest)
	case "mining":
		return parseMiningChange(lineNum, rest)
	case "transfer":
		return parseTransfer(lineNum, rest)
	case "assemble":
		return parseAssemble(lineNum, rest)
	case "move":
		return parseMove(lineNum, rest)
	case "draft":
		return parseDraft(lineNum, rest)
	case "pay":
		return parsePay(lineNum, rest)
	case "ration":
		return parseRation(lineNum, rest)
	case "name":
		return parseNameByKind(lineNum, rest)
	// known non-MVP commands
	case "bombard", "invade", "raid", "support", "disassemble",
		"buy", "sell", "survey", "probe", "disband", "control",
		"permission", "news", "check", "convert", "incite",
		"attack", "gather", "un-control":
		return nil, notImplemented(lineNum, keyword), nil
	default:
		return nil, unknownCommand(lineNum, keyword), nil
	}
}

// parseBuildChange parses: build change <colonyID> <groupNo> <unitKind>
func parseBuildChange(lineNum int, rest []string) (domain.Order, *app.ParseDiagnostic, error) {
	if len(rest) < 1 || strings.ToLower(rest[0]) != "change" {
		return nil, badSyntax(lineNum, "expected 'build change <colonyID> <groupNo> <unitKind>'"), nil
	}
	rest = rest[1:]

	if len(rest) < 3 {
		return nil, badSyntax(lineNum, "build change requires <colonyID> <groupNo> <unitKind>"), nil
	}

	colonyID, err := parsePositiveInt(rest[0])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("build change: invalid colony ID %q: %v", rest[0], err)), nil
	}

	groupNo, err := parsePositiveInt(rest[1])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("build change: invalid group number %q: %v", rest[1], err)), nil
	}

	unitKind, err := parseUnitKind(rest[2])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("build change: invalid unit kind %q", rest[2])), nil
	}

	o := domain.BuildChangeOrder{
		OrderKind:      domain.OrderKindBuildChange,
		ColonyID:       domain.ColonyID(colonyID),
		FactoryGroupID: domain.FactoryGroupID(groupNo),
		NewUnitKind:    unitKind,
	}
	if err := o.Validate(); err != nil {
		return nil, badValue(lineNum, err.Error()), nil
	}
	return o, nil, nil
}

// parseMiningChange parses: mining change <colonyID> <groupNo> <depositID>
func parseMiningChange(lineNum int, rest []string) (domain.Order, *app.ParseDiagnostic, error) {
	if len(rest) < 1 || strings.ToLower(rest[0]) != "change" {
		return nil, badSyntax(lineNum, "expected 'mining change <colonyID> <groupNo> <depositID>'"), nil
	}
	rest = rest[1:]

	if len(rest) < 3 {
		return nil, badSyntax(lineNum, "mining change requires <colonyID> <groupNo> <depositID>"), nil
	}

	colonyID, err := parsePositiveInt(rest[0])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("mining change: invalid colony ID %q: %v", rest[0], err)), nil
	}

	groupNo, err := parsePositiveInt(rest[1])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("mining change: invalid group number %q: %v", rest[1], err)), nil
	}

	depositID, err := parsePositiveInt(rest[2])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("mining change: invalid deposit ID %q: %v", rest[2], err)), nil
	}

	o := domain.MiningChangeOrder{
		OrderKind:     domain.OrderKindMiningChange,
		ColonyID:      domain.ColonyID(colonyID),
		MiningGroupID: domain.MiningGroupID(groupNo),
		DepositID:     domain.DepositID(depositID),
	}
	if err := o.Validate(); err != nil {
		return nil, badValue(lineNum, err.Error()), nil
	}
	return o, nil, nil
}

// parseTransfer parses: transfer <sourceID> <destID> <unitKind> <qty>
func parseTransfer(lineNum int, rest []string) (domain.Order, *app.ParseDiagnostic, error) {
	if len(rest) < 4 {
		return nil, badSyntax(lineNum, "transfer requires <sourceID> <destID> <unitKind> <qty>"), nil
	}

	sourceID, err := parsePositiveInt(rest[0])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("transfer: invalid source ID %q: %v", rest[0], err)), nil
	}

	destID, err := parsePositiveInt(rest[1])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("transfer: invalid destination ID %q: %v", rest[1], err)), nil
	}

	unitKind, err := parseUnitKind(rest[2])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("transfer: invalid unit kind %q", rest[2])), nil
	}

	quantity, err := parsePositiveInt(rest[3])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("transfer: invalid quantity %q: %v", rest[3], err)), nil
	}

	o := domain.TransferOrder{
		OrderKind: domain.OrderKindTransfer,
		SourceID:  domain.ColonyID(sourceID),
		DestID:    domain.ColonyID(destID),
		UnitKind:  unitKind,
		Quantity:  quantity,
	}
	if err := o.Validate(); err != nil {
		return nil, badValue(lineNum, err.Error()), nil
	}
	return o, nil, nil
}

// parseAssemble parses: assemble <locationID> <unitKind> <qty>
func parseAssemble(lineNum int, rest []string) (domain.Order, *app.ParseDiagnostic, error) {
	if len(rest) < 3 {
		return nil, badSyntax(lineNum, "assemble requires <locationID> <unitKind> <qty>"), nil
	}

	locationID, err := parsePositiveInt(rest[0])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("assemble: invalid location ID %q: %v", rest[0], err)), nil
	}

	unitKind, err := parseUnitKind(rest[1])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("assemble: invalid unit kind %q", rest[1])), nil
	}

	quantity, err := parsePositiveInt(rest[2])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("assemble: invalid quantity %q: %v", rest[2], err)), nil
	}

	o := domain.AssembleOrder{
		OrderKind: domain.OrderKindAssemble,
		ColonyID:  domain.ColonyID(locationID),
		UnitKind:  unitKind,
		Quantity:  quantity,
	}
	if err := o.Validate(); err != nil {
		return nil, badValue(lineNum, err.Error()), nil
	}
	return o, nil, nil
}

// parseMove parses: move <shipID> <destination...>
func parseMove(lineNum int, rest []string) (domain.Order, *app.ParseDiagnostic, error) {
	if len(rest) < 2 {
		return nil, badSyntax(lineNum, "move requires <shipID> <destination>"), nil
	}

	shipID, err := parsePositiveInt(rest[0])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("move: invalid ship ID %q: %v", rest[0], err)), nil
	}

	destination := strings.Join(rest[1:], " ")
	o := domain.MoveOrder{
		OrderKind:   domain.OrderKindMove,
		ShipID:      domain.ShipID(shipID),
		Destination: domain.MoveDestination{Raw: destination},
	}
	if err := o.Validate(); err != nil {
		return nil, badValue(lineNum, err.Error()), nil
	}
	return o, nil, nil
}

// parseDraft parses: draft <colonyID> <popKind> <qty>
func parseDraft(lineNum int, rest []string) (domain.Order, *app.ParseDiagnostic, error) {
	if len(rest) < 3 {
		return nil, badSyntax(lineNum, "draft requires <colonyID> <popKind> <qty>"), nil
	}

	colonyID, err := parsePositiveInt(rest[0])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("draft: invalid colony ID %q: %v", rest[0], err)), nil
	}

	popKind, err := parseUnitKind(rest[1])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("draft: invalid population kind %q", rest[1])), nil
	}

	quantity, err := parsePositiveInt(rest[2])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("draft: invalid quantity %q: %v", rest[2], err)), nil
	}

	o := domain.DraftOrder{
		OrderKind: domain.OrderKindDraft,
		ColonyID:  domain.ColonyID(colonyID),
		PopKind:   popKind,
		Quantity:  quantity,
	}
	if err := o.Validate(); err != nil {
		return nil, badValue(lineNum, err.Error()), nil
	}
	return o, nil, nil
}

// parsePay parses: pay <colonyID> <popKind> <amount>
func parsePay(lineNum int, rest []string) (domain.Order, *app.ParseDiagnostic, error) {
	if len(rest) < 3 {
		return nil, badSyntax(lineNum, "pay requires <colonyID> <popKind> <amount>"), nil
	}

	colonyID, err := parsePositiveInt(rest[0])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("pay: invalid colony ID %q: %v", rest[0], err)), nil
	}

	popKind, err := parseUnitKind(rest[1])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("pay: invalid population kind %q", rest[1])), nil
	}

	amount, err := parseNonNegativeInt(rest[2])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("pay: invalid wage amount %q: %v", rest[2], err)), nil
	}

	o := domain.PayOrder{
		OrderKind: domain.OrderKindPay,
		ColonyID:  domain.ColonyID(colonyID),
		PopKind:   popKind,
		Wage:      amount,
	}
	if err := o.Validate(); err != nil {
		return nil, badValue(lineNum, err.Error()), nil
	}
	return o, nil, nil
}

// parseRation parses: ration <colonyID> <percentage>
func parseRation(lineNum int, rest []string) (domain.Order, *app.ParseDiagnostic, error) {
	if len(rest) < 2 {
		return nil, badSyntax(lineNum, "ration requires <colonyID> <percentage>"), nil
	}

	colonyID, err := parsePositiveInt(rest[0])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("ration: invalid colony ID %q: %v", rest[0], err)), nil
	}

	pct, err := parsePercentage(rest[1])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("ration: invalid percentage %q: %v", rest[1], err)), nil
	}

	o := domain.RationOrder{
		OrderKind:        domain.OrderKindRation,
		ColonyID:         domain.ColonyID(colonyID),
		RationPercentage: pct,
	}
	if err := o.Validate(); err != nil {
		return nil, badValue(lineNum, err.Error()), nil
	}
	return o, nil, nil
}

// parseNameByKind parses: name planet|ship|colony <id> <newName>
func parseNameByKind(lineNum int, rest []string) (domain.Order, *app.ParseDiagnostic, error) {
	if len(rest) < 3 {
		return nil, badSyntax(lineNum, "name requires <planet|ship|colony> <id> <newName>"), nil
	}

	var targetKind domain.NameTargetKind
	switch strings.ToLower(rest[0]) {
	case "planet":
		targetKind = domain.NameTargetPlanet
	case "ship":
		targetKind = domain.NameTargetShip
	case "colony":
		targetKind = domain.NameTargetColony
	default:
		return nil, badSyntax(lineNum, fmt.Sprintf("name: unknown target kind %q", rest[0])), nil
	}

	targetID, err := parsePositiveInt(rest[1])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("name: invalid target ID %q: %v", rest[1], err)), nil
	}

	newName := strings.Join(rest[2:], " ")
	o := domain.NameOrder{
		OrderKind:  domain.OrderKindName,
		TargetKind: targetKind,
		TargetID:   targetID,
		NewName:    newName,
	}
	if err := o.Validate(); err != nil {
		return nil, badValue(lineNum, err.Error()), nil
	}
	return o, nil, nil
}

func parsePositiveInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("not an integer")
	}
	if n <= 0 {
		return 0, fmt.Errorf("must be positive, got %d", n)
	}
	return n, nil
}

func parseNonNegativeInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("not an integer")
	}
	if n < 0 {
		return 0, fmt.Errorf("must be non-negative, got %d", n)
	}
	return n, nil
}

func parsePercentage(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("not an integer")
	}
	if n < 0 || n > 100 {
		return 0, fmt.Errorf("percentage must be 0–100, got %d", n)
	}
	return n, nil
}

// parseUnitKind maps a unit token string to a domain.UnitKind.
func parseUnitKind(s string) (domain.UnitKind, error) {
	switch strings.ToLower(s) {
	case "unemployable", "unemployables":
		return domain.Unemployables, nil
	case "unskilled-worker", "unskilled-workers", "unskilled":
		return domain.UnskilledWorkers, nil
	case "professional", "professionals", "pro":
		return domain.Professionals, nil
	case "soldier", "soldiers":
		return domain.Soldiers, nil
	case "spy", "spies":
		return domain.Spies, nil
	case "construction-worker", "construction-workers":
		return domain.ConstructionWorkers, nil
	case "rebel", "rebels":
		return domain.Rebels, nil
	case "assault-craft":
		return domain.AssaultCraft, nil
	case "assault-weapon":
		return domain.AssaultWeapon, nil
	case "anti-missile":
		return domain.AntiMissile, nil
	case "energy-shield":
		return domain.EnergyShield, nil
	case "energy-weapon":
		return domain.EnergyWeapon, nil
	case "military-robot":
		return domain.MilitaryRobot, nil
	case "military-supply":
		return domain.MilitarySupply, nil
	case "missile":
		return domain.Missile, nil
	case "missile-launcher":
		return domain.MissileLauncher, nil
	case "farm":
		return domain.Farm, nil
	case "factory", "fact":
		return domain.Factory, nil
	case "mine":
		return domain.Mine, nil
	case "automation":
		return domain.Automation, nil
	case "consumer-goods", "cons":
		return domain.ConsumerGoods, nil
	case "food":
		return domain.Food, nil
	case "hyper-engine", "hype":
		return domain.HyperEngine, nil
	case "life-support":
		return domain.LifeSupport, nil
	case "light-structural":
		return domain.LightStructural, nil
	case "sensor", "sen":
		return domain.Sensor, nil
	case "space-drive", "spac":
		return domain.SpaceDrive, nil
	case "structural":
		return domain.Structural, nil
	case "transport":
		return domain.Transport, nil
	case "research-point":
		return domain.ResearchPoint, nil
	default:
		return 0, fmt.Errorf("unknown unit kind %q", s)
	}
}

func notImplemented(line int, cmd string) *app.ParseDiagnostic {
	return &app.ParseDiagnostic{
		Line:    line,
		Code:    "not_implemented",
		Message: fmt.Sprintf("%s is not yet implemented", cmd),
	}
}

func unknownCommand(line int, cmd string) *app.ParseDiagnostic {
	return &app.ParseDiagnostic{
		Line:    line,
		Code:    "unknown_command",
		Message: fmt.Sprintf("unrecognized command %q", cmd),
	}
}

func badSyntax(line int, msg string) *app.ParseDiagnostic {
	return &app.ParseDiagnostic{
		Line:    line,
		Code:    "bad_syntax",
		Message: msg,
	}
}

func badValue(line int, msg string) *app.ParseDiagnostic {
	return &app.ParseDiagnostic{
		Line:    line,
		Code:    "bad_value",
		Message: msg,
	}
}
