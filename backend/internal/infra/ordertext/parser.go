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

		// strip comment: // starts a comment outside a quoted string
		if idx := commentIndex(raw); idx >= 0 {
			raw = raw[:idx]
		}

		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		// detect unterminated quoted string before tokenizing
		if hasUnterminatedQuote(line) {
			diags = append(diags, *unterminatedQuote(lineNum))
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
	case "end":
		return nil, unexpectedEnd(lineNum), nil
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

	quantity, err := parseQuantity(rest[3])
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

// parseAssemble dispatches to the correct assemble form based on the second token:
//
//	factory → assemble <id> factory <factory-unit> <qty> <build-target>
//	mine    → assemble <id> mine <mine-unit> <qty> <deposit-id>
//	other   → assemble <id> <unit-token> <qty>
func parseAssemble(lineNum int, rest []string) (domain.Order, *app.ParseDiagnostic, error) {
	if len(rest) < 3 {
		return nil, badSyntax(lineNum, "assemble requires <locationID> <unit> <qty>"), nil
	}

	locationID, err := parsePositiveInt(rest[0])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("assemble: invalid location ID %q: %v", rest[0], err)), nil
	}

	switch strings.ToLower(rest[1]) {
	case "factory":
		return parseAssembleFactory(lineNum, locationID, rest[2:])
	case "mine":
		return parseAssembleMine(lineNum, locationID, rest[2:])
	default:
		return parseAssembleOther(lineNum, locationID, rest[1:])
	}
}

// parseAssembleOther parses: assemble <id> <unit-token> <qty>
// rest = [<unit-token>, <qty>]
func parseAssembleOther(lineNum int, locationID int, rest []string) (domain.Order, *app.ParseDiagnostic, error) {
	if len(rest) < 2 {
		return nil, badSyntax(lineNum, "assemble requires <locationID> <unit> <qty>"), nil
	}

	unitKind, err := parseUnitKind(rest[0])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("assemble: invalid unit kind %q: %v", rest[0], err)), nil
	}

	quantity, err := parseQuantity(rest[1])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("assemble: invalid quantity %q: %v", rest[1], err)), nil
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

// parseAssembleFactory parses: assemble <id> factory <factory-unit> <qty> <build-target>
// rest = [<factory-unit>, <qty>, <build-target>]
func parseAssembleFactory(lineNum int, locationID int, rest []string) (domain.Order, *app.ParseDiagnostic, error) {
	if len(rest) < 3 {
		return nil, badSyntax(lineNum, "assemble factory requires <factory-unit> <qty> <build-target>"), nil
	}

	factoryUnit, err := parseUnitKind(rest[0])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("assemble factory: invalid factory unit %q: %v", rest[0], err)), nil
	}

	quantity, err := parseQuantity(rest[1])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("assemble factory: invalid quantity %q: %v", rest[1], err)), nil
	}

	buildTarget, err := parseUnitKind(rest[2])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("assemble factory: invalid build target %q: %v", rest[2], err)), nil
	}

	o := domain.AssembleFactoryOrder{
		OrderKind:   domain.OrderKindAssemble,
		LocationID:  domain.ColonyID(locationID),
		FactoryUnit: factoryUnit,
		FactoryQty:  quantity,
		BuildTarget: buildTarget,
	}
	if err := o.Validate(); err != nil {
		return nil, badValue(lineNum, err.Error()), nil
	}
	return o, nil, nil
}

// parseAssembleMine parses: assemble <id> mine <mine-unit> <qty> <deposit-id>
// rest = [<mine-unit>, <qty>, <deposit-id>]
func parseAssembleMine(lineNum int, locationID int, rest []string) (domain.Order, *app.ParseDiagnostic, error) {
	if len(rest) < 3 {
		return nil, badSyntax(lineNum, "assemble mine requires <mine-unit> <qty> <deposit-id>"), nil
	}

	mineUnit, err := parseUnitKind(rest[0])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("assemble mine: invalid mine unit %q: %v", rest[0], err)), nil
	}

	quantity, err := parseQuantity(rest[1])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("assemble mine: invalid quantity %q: %v", rest[1], err)), nil
	}

	depositID, err := parsePositiveInt(rest[2])
	if err != nil {
		return nil, badValue(lineNum, fmt.Sprintf("assemble mine: invalid deposit ID %q: %v", rest[2], err)), nil
	}

	o := domain.AssembleMineOrder{
		OrderKind:  domain.OrderKindAssemble,
		LocationID: domain.ColonyID(locationID),
		MineUnit:   mineUnit,
		MineQty:    quantity,
		DepositID:  domain.DepositID(depositID),
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

	quantity, err := parseQuantity(rest[2])
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

	// Name must be a quoted string. Joining handles multi-word names that
	// strings.Fields split across tokens (e.g. "New Terra" → ["New", "Terra"]).
	raw := strings.Join(rest[2:], " ")
	if !strings.HasPrefix(raw, `"`) || !strings.HasSuffix(raw, `"`) {
		return nil, badSyntax(lineNum, `name: name must be a quoted string (e.g. "Dragonfire")`), nil
	}
	newName := scrubName(raw[1 : len(raw)-1]) // strip quotes then scrub

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

// parseQuantity strips comma thousands-separators then parses as a positive integer.
// "54,000" and "54000" are equivalent.
func parseQuantity(s string) (int, error) {
	return parsePositiveInt(strings.ReplaceAll(s, ",", ""))
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

// unitEntry records the domain kind for a canonical base token and whether a
// tech-level suffix (e.g. "-1") is required.
type unitEntry struct {
	kind         domain.UnitKind
	requiresTech bool
}

// unitRegistry maps canonical lowercase base tokens to their entry.
// Equipment units require a tech-level suffix; population and commodities do not.
var unitRegistry = map[string]unitEntry{
	// population — no tech level
	"unemployable":         {domain.Unemployables, false},
	"unemployables":        {domain.Unemployables, false},
	"unskilled-worker":     {domain.UnskilledWorkers, false},
	"unskilled-workers":    {domain.UnskilledWorkers, false},
	"unskilled":            {domain.UnskilledWorkers, false},
	"professional":         {domain.Professionals, false},
	"professionals":        {domain.Professionals, false},
	"pro":                  {domain.Professionals, false},
	"soldier":              {domain.Soldiers, false},
	"soldiers":             {domain.Soldiers, false},
	"spy":                  {domain.Spies, false},
	"spies":                {domain.Spies, false},
	"construction-worker":  {domain.ConstructionWorkers, false},
	"construction-workers": {domain.ConstructionWorkers, false},
	"rebel":                {domain.Rebels, false},
	"rebels":               {domain.Rebels, false},
	// commodities — no tech level
	"consumer-goods":  {domain.ConsumerGoods, false},
	"cons":            {domain.ConsumerGoods, false},
	"food":            {domain.Food, false},
	"structural":      {domain.Structural, false},
	"light-structural": {domain.LightStructural, false},
	"research-point":  {domain.ResearchPoint, false},
	// equipment — tech level required
	"factory":          {domain.Factory, true},
	"fact":             {domain.Factory, true},
	"mine":             {domain.Mine, true},
	"farm":             {domain.Farm, true},
	"hyper-engine":     {domain.HyperEngine, true},
	"hype":             {domain.HyperEngine, true},
	"space-drive":      {domain.SpaceDrive, true},
	"spac":             {domain.SpaceDrive, true},
	"life-support":     {domain.LifeSupport, true},
	"sensor":           {domain.Sensor, true},
	"sen":              {domain.Sensor, true},
	"automation":       {domain.Automation, true},
	"transport":        {domain.Transport, true},
	"energy-weapon":    {domain.EnergyWeapon, true},
	"energy-shield":    {domain.EnergyShield, true},
	"anti-missile":     {domain.AntiMissile, true},
	"assault-craft":    {domain.AssaultCraft, true},
	"assault-weapon":   {domain.AssaultWeapon, true},
	"military-robot":   {domain.MilitaryRobot, true},
	"military-supply":  {domain.MilitarySupply, true},
	"missile":          {domain.Missile, true},
	"missile-launcher": {domain.MissileLauncher, true},
}

// parseUnitKind parses a unit token such as "factory-6" or "professional".
// Equipment units require a "-N" tech-level suffix (e.g. factory-6, hyper-engine-1).
// Population and commodity units must not carry a tech-level suffix.
func parseUnitKind(s string) (domain.UnitKind, error) {
	base, techLevel := splitTechLevel(strings.ToLower(s))
	hasTech := techLevel > 0

	entry, ok := unitRegistry[base]
	if !ok {
		return 0, fmt.Errorf("unknown unit kind %q", s)
	}
	if entry.requiresTech && !hasTech {
		return 0, fmt.Errorf("%q requires a tech-level suffix (e.g. %s-1)", s, base)
	}
	if !entry.requiresTech && hasTech {
		return 0, fmt.Errorf("%q does not use a tech-level suffix", s)
	}
	return entry.kind, nil
}

// splitTechLevel splits a trailing "-<positive-integer>" suffix from s.
// Returns the base string and tech level. Tech level is 0 when absent.
func splitTechLevel(s string) (base string, techLevel int) {
	idx := strings.LastIndex(s, "-")
	if idx < 0 || idx == len(s)-1 {
		return s, 0
	}
	n, err := strconv.Atoi(s[idx+1:])
	if err != nil || n <= 0 {
		return s, 0
	}
	return s[:idx], n
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
		Code:    "syntax",
		Message: msg,
	}
}

func badValue(line int, msg string) *app.ParseDiagnostic {
	return &app.ParseDiagnostic{
		Line:    line,
		Code:    "invalid_value",
		Message: msg,
	}
}

func unterminatedQuote(line int) *app.ParseDiagnostic {
	return &app.ParseDiagnostic{
		Line:    line,
		Code:    "unterminated_quote",
		Message: "unterminated quoted string",
	}
}

func unexpectedEnd(line int) *app.ParseDiagnostic {
	return &app.ParseDiagnostic{
		Line:    line,
		Code:    "unexpected_end",
		Message: "end without an open setup block",
	}
}

// scrubName normalises a name by removing any character that is not an ASCII
// letter, ASCII digit, space, or an allowed punctuation character, then
// trimming leading/trailing whitespace and collapsing internal runs of
// whitespace to a single space.
//
// Allowed punctuation: space / , . # - _ + ( ) '
func scrubName(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ', r == '/', r == ',', r == '.', r == '#',
			r == '-', r == '_', r == '+', r == '(', r == ')', r == '\'':
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// hasUnterminatedQuote reports whether s contains an opening double-quote with
// no matching closing quote. It assumes comments have already been stripped.
func hasUnterminatedQuote(s string) bool {
	count := 0
	for _, c := range s {
		if c == '"' {
			count++
		}
	}
	return count%2 != 0
}

// commentIndex returns the index of the first // that appears outside a quoted
// string, or -1 if there is none. Quotes inside // comments are not tracked.
func commentIndex(s string) int {
	inQuote := false
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			inQuote = !inQuote
		case '/':
			if !inQuote && i+1 < len(s) && s[i+1] == '/' {
				return i
			}
		}
	}
	return -1
}
