// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package clustergen

import (
	"fmt"
	"slices"
	"sort"

	"github.com/mdhender/ec/internal/domain"
	"github.com/mdhender/prng"
)

// Internal tree types used during generation only.

type system struct {
	location domain.Coords
	stars    []*star
}

type star struct {
	location domain.Coords
	sequence int
	planets  []*planet
}

type planet struct {
	kind         domain.PlanetKind
	orbit        int // 1...10
	habitability int // 0...25
	deposits     []*deposit
}

type deposit struct {
	kind     domain.NaturalResource
	quantity int // 1,000,000 to 99,999,999
	yieldPct int // 1 to 100
}

func (d *deposit) less(d2 *deposit) bool {
	if d.kind == d2.kind {
		if d.yieldPct == d2.yieldPct {
			return d.quantity < d2.quantity
		}
		return d.yieldPct < d2.yieldPct
	}
	return d.kind < d2.kind
}

// GenerateCluster generates a complete cluster and returns it as a
// normalized domain.Cluster with shuffled IDs.
func GenerateCluster(r *prng.Rand) (domain.Cluster, error) {
	systems, err := generateSystems(r)
	if err != nil {
		return domain.Cluster{}, err
	}
	return normalize(r, systems), nil
}

// normalize converts the internal tree representation into a flat,
// ID-referenced domain.Cluster. Planet and Deposit IDs are shuffled
// so players cannot infer counts from sequential numbering.
func normalize(r *prng.Rand, systems []*system) domain.Cluster {
	var c domain.Cluster

	// Count total planets and deposits across all systems.
	totalPlanets := 0
	totalDeposits := 0
	for _, gs := range systems {
		for _, gStar := range gs.stars {
			totalPlanets += len(gStar.planets)
			for _, gPlanet := range gStar.planets {
				totalDeposits += len(gPlanet.deposits)
			}
		}
	}

	// Generate planet IDs 1..N and shuffle them.
	planetIDs := make([]domain.PlanetID, totalPlanets)
	for i := range planetIDs {
		planetIDs[i] = domain.PlanetID(i + 1)
	}
	r.Shuffle(len(planetIDs), func(i, j int) {
		planetIDs[i], planetIDs[j] = planetIDs[j], planetIDs[i]
	})

	// Generate deposit IDs 1..N and shuffle them.
	depositIDs := make([]domain.DepositID, totalDeposits)
	for i := range depositIDs {
		depositIDs[i] = domain.DepositID(i + 1)
	}
	r.Shuffle(len(depositIDs), func(i, j int) {
		depositIDs[i], depositIDs[j] = depositIDs[j], depositIDs[i]
	})

	nextSystemID := domain.SystemID(1)
	nextStarID := domain.StarID(1)
	planetIdx := 0
	depIdx := 0

	// Map from DepositID to its index in c.Deposits for sorting.
	depositIndex := make(map[domain.DepositID]int, totalDeposits)

	for _, gs := range systems {
		sysID := nextSystemID
		nextSystemID++

		sys := domain.System{
			ID:       sysID,
			Location: gs.location,
			Display:  fmt.Sprintf("%02d-%02d-%02d", gs.location.X, gs.location.Y, gs.location.Z),
		}

		for _, gStar := range gs.stars {
			starID := nextStarID
			nextStarID++

			dStar := domain.Star{
				ID:       starID,
				Sequence: gStar.sequence,
				System:   sysID,
			}
			if len(gs.stars) == 1 {
				dStar.Display = sys.Display
			} else {
				dStar.Display = fmt.Sprintf("%s/%d", sys.Display, gStar.sequence)
			}

			for _, gPlanet := range gStar.planets {
				planetID := planetIDs[planetIdx]
				planetIdx++

				dPlanet := domain.Planet{
					ID:           planetID,
					Kind:         gPlanet.kind,
					Habitability: gPlanet.habitability,
				}

				for _, gDep := range gPlanet.deposits {
					depID := depositIDs[depIdx]
					depIdx++

					dep := domain.Deposit{
						ID:                depID,
						Resource:          gDep.kind,
						YieldPct:          gDep.yieldPct,
						QuantityRemaining: gDep.quantity,
					}
					c.Deposits = append(c.Deposits, dep)
					depositIndex[depID] = len(c.Deposits) - 1
					dPlanet.Deposits = append(dPlanet.Deposits, depID)
				}

				// Sort planet's deposit IDs by Kind, then YieldPct, then Quantity.
				sort.Slice(dPlanet.Deposits, func(i, j int) bool {
					di := c.Deposits[depositIndex[dPlanet.Deposits[i]]]
					dj := c.Deposits[depositIndex[dPlanet.Deposits[j]]]
					if di.Resource != dj.Resource {
						return di.Resource < dj.Resource
					}
					if di.YieldPct != dj.YieldPct {
						return di.YieldPct < dj.YieldPct
					}
					return di.QuantityRemaining < dj.QuantityRemaining
				})

				dStar.Orbits[gPlanet.orbit-1] = planetID
				c.Planets = append(c.Planets, dPlanet)
			}

			sys.Stars = append(sys.Stars, starID)
			c.Stars = append(c.Stars, dStar)
		}

		c.Systems = append(c.Systems, sys)
	}

	// Sort flat slices by ID so lookups are predictable.
	sort.Slice(c.Planets, func(i, j int) bool {
		return c.Planets[i].ID < c.Planets[j].ID
	})
	sort.Slice(c.Deposits, func(i, j int) bool {
		return c.Deposits[i].ID < c.Deposits[j].ID
	})

	return c
}

// generateSystems returns a slice of systems containing a total of 100 stars.
func generateSystems(r *prng.Rand) ([]*system, error) {
	pts, err := generatePoints(r)
	if err != nil {
		return nil, err
	}

	systems := make([]*system, 0, 100)
	var s *system
	for _, pt := range pts {
		if s == nil || s.location != pt {
			s = &system{location: pt}
			systems = append(systems, s)
		}
	}

	systemNo := 0
	for _, pt := range pts {
		if systems[systemNo].location != pt {
			systemNo++
		}
		s = systems[systemNo]
		st, err := generateStar(r, pt, len(s.stars))
		if err != nil {
			return nil, err
		}
		s.stars = append(s.stars, st)
	}

	return systems, nil
}

// generatePoints returns a deterministic set of 100 coordinates within a
// 31×31×31 cube using pseudo-random sampling.
//
// Each coordinate component (X, Y, Z) is drawn independently from a discrete
// uniform distribution over [0, 30]. As a result:
//
//   - Points are uniformly distributed across the cube
//   - Duplicate coordinates may occur
//   - No minimum spacing or clustering is enforced
//
// The resulting slice is sorted in lexicographic order (X → Y → Z) to ensure
// stable, reproducible output for a given PRNG seed. This is important for
// golden files, testing, and deterministic world generation.
//
// Note: This function does not generate a statistical "cluster" (e.g.,
// Gaussian, Plummer, or Poisson disk). It implements the simple uniform
// model used in early versions of the game. Future versions may introduce
// alternative distributions or configurable generation models.
func generatePoints(r *prng.Rand) ([]domain.Coords, error) {
	c := make([]domain.Coords, 100)
	for n := 0; n < 100; n++ {
		c[n] = domain.Coords{
			X: r.IntN(31),
			Y: r.IntN(31),
			Z: r.IntN(31),
		}
	}

	sort.Slice(c, func(i, j int) bool {
		return c[i].Less(c[j])
	})

	return c, nil
}

func generateStar(r *prng.Rand, pt domain.Coords, seq int) (*star, error) {
	s := &star{location: pt, sequence: seq, planets: make([]*planet, 0, 10)}
	orbits, err := generateOrbits(r)
	if err != nil {
		return nil, err
	}
	for o, k := range orbits {
		var p *planet
		switch k {
		case domain.AsteroidBelt:
			p, err = generateAsteroidBelt(r, o+1)
			if err != nil {
				return nil, err
			}
		case domain.GasGiant:
			p, err = generateGasGiant(r, o+1)
			if err != nil {
				return nil, err
			}
		case domain.Terrestrial:
			p, err = generateTerrestrial(r, o+1)
			if err != nil {
				return nil, err
			}
		default:
			continue
		}
		sort.Slice(p.deposits, func(i, j int) bool {
			return p.deposits[i].less(p.deposits[j])
		})
		s.planets = append(s.planets, p)
	}
	return s, nil
}

func generateOrbits(r *prng.Rand) ([]domain.PlanetKind, error) {
	o := make([]domain.PlanetKind, 10)
	emptyOrbits := 0
	for n := 0; n < 10; n++ {
		switch roll := r.IntN(100) + 1; {
		case roll <= 29:
			o[n] = domain.Terrestrial
		case roll <= 34:
			o[n] = domain.AsteroidBelt
		case roll <= 41:
			o[n] = domain.GasGiant
		default:
			emptyOrbits++
		}
	}
	if emptyOrbits == 10 {
		o = []domain.PlanetKind{
			domain.Terrestrial,
			domain.Terrestrial,
			domain.Terrestrial,
			domain.Terrestrial,
			domain.AsteroidBelt,
			domain.GasGiant,
			domain.GasGiant,
			domain.GasGiant,
			domain.Terrestrial,
			domain.AsteroidBelt,
		}
	}
	return finalizeOrbits(o)
}

func finalizeOrbits(orbits []domain.PlanetKind) ([]domain.PlanetKind, error) {
	o := make([]domain.PlanetKind, 10)
	for orbit, k := range orbits {
		o[orbit] = k
	}

	gasGiants := countOrbits(o, domain.GasGiant)
	for orbit := 0; orbit < 10 && gasGiants > 3; orbit++ {
		if o[orbit] == domain.GasGiant {
			o[orbit] = domain.AsteroidBelt
			gasGiants--
		}
	}

	asteroids := countOrbits(o, domain.AsteroidBelt)
	for orbit := 0; orbit < 10 && asteroids > 2; orbit++ {
		if o[orbit] == domain.AsteroidBelt {
			o[orbit] = domain.Terrestrial
			asteroids--
		}
	}

	var occupied []int
	var types []domain.PlanetKind
	for orbit := 0; orbit < 10; orbit++ {
		switch o[orbit] {
		case domain.AsteroidBelt:
			occupied = append(occupied, orbit)
			types = append(types, o[orbit])
		case domain.GasGiant:
			occupied = append(occupied, orbit)
			types = append(types, o[orbit])
		case domain.Terrestrial:
			occupied = append(occupied, orbit)
			types = append(types, o[orbit])
		}
	}
	slices.Sort(types)
	for i, orbit := range occupied {
		o[orbit] = types[i]
	}

	return o, nil
}

func countOrbits(orbits []domain.PlanetKind, want domain.PlanetKind) int {
	count := 0
	for _, k := range orbits {
		if k == want {
			count++
		}
	}
	return count
}

func generateAsteroidBelt(r *prng.Rand, orbit int) (*planet, error) {
	p := &planet{kind: domain.AsteroidBelt,
		orbit:        orbit,
		habitability: 0,
		deposits:     make([]*deposit, r.IntN(35)),
	}
	var err error
	for d := range p.deposits {
		p.deposits[d], err = generateAsteroidBeltDeposits(r)
		if err != nil {
			return nil, err
		}
	}
	return p, nil
}

func generateGasGiant(r *prng.Rand, orbit int) (*planet, error) {
	p := &planet{kind: domain.GasGiant,
		orbit:    orbit,
		deposits: make([]*deposit, r.IntN(35)),
	}
	switch orbit {
	case 1:
		p.habitability = r.IntN(3)
	case 2:
		p.habitability = r.IntN(5)
	case 3:
		p.habitability = r.IntN(8)
	case 4:
		p.habitability = r.IntN(11)
	case 5:
		p.habitability = r.IntN(14)
	case 6:
		p.habitability = r.IntN(17)
	case 7:
		p.habitability = r.IntN(20)
	case 8:
		p.habitability = r.IntN(14)
	case 9:
		p.habitability = r.IntN(8)
	case 10:
		p.habitability = r.IntN(2)
	default:
		return nil, fmt.Errorf("invalid orbit %d", orbit)
	}
	var err error
	for d := range p.deposits {
		p.deposits[d], err = generateGasGiantDeposits(r)
		if err != nil {
			return nil, err
		}
	}
	return p, nil
}

func generateTerrestrial(r *prng.Rand, orbit int) (*planet, error) {
	p := &planet{kind: domain.Terrestrial,
		orbit:    orbit,
		deposits: make([]*deposit, r.IntN(35)),
	}
	switch orbit {
	case 1:
		p.habitability = rollDice(r, 1, 3) - 1
	case 2:
		p.habitability = rollDice(r, 2, 3) - 2
	case 3:
		p.habitability = rollDice(r, 3, 4) - 3
	case 4:
		p.habitability = rollDice(r, 2, 10)
	case 5:
		p.habitability = rollDice(r, 2, 12) + 1
	case 6:
		p.habitability = rollDice(r, 2, 10)
	case 7:
		p.habitability = rollDice(r, 3, 4) - 3
	case 8:
		p.habitability = rollDice(r, 2, 3) - 2
	case 9:
		p.habitability = rollDice(r, 1, 3) - 1
	case 10:
		p.habitability = rollDice(r, 1, 2) - 1
	default:
		return nil, fmt.Errorf("invalid orbit %d", orbit)
	}
	var err error
	for d := range p.deposits {
		p.deposits[d], err = generateTerrestrialDeposits(r, p.habitability)
		if err != nil {
			return nil, err
		}
	}
	return p, nil
}

func generateAsteroidBeltDeposits(r *prng.Rand) (*deposit, error) {
	roll := r.IntN(100) + 1 // 1...100
	switch {
	case roll == 1:
		return &deposit{kind: domain.GOLD,
			quantity: r.IntN(4_900_001) + 100_000, // 100,000 to 5,000,000 inclusive
			yieldPct: rollDice(r, 1, 3),
		}, nil
	case roll <= 10:
		return &deposit{kind: domain.FUEL,
			quantity: r.IntN(98_000_001) + 1_000_000, // 1,000,000 to 99,000,000 inclusive
			yieldPct: rollDice(r, 3, 6) - 2,
		}, nil
	default:
		return &deposit{kind: domain.METALLICS,
			quantity: r.IntN(98_000_001) + 1_000_000, // 1,000,000 to 99,000,000 inclusive
			yieldPct: rollDice(r, 3, 10) - 2,
		}, nil
	}
}

func generateGasGiantDeposits(r *prng.Rand) (*deposit, error) {
	roll := r.IntN(100) + 1 // 1...100
	switch {
	case roll <= 15:
		return &deposit{kind: domain.FUEL,
			quantity: r.IntN(98_000_001) + 1_000_000, // 1,000,000 to 99,000,000 inclusive
			yieldPct: rollDice(r, 10, 4) - 2,
		}, nil
	case roll <= 40:
		return &deposit{kind: domain.METALLICS,
			quantity: r.IntN(98_000_001) + 1_000_000, // 1,000,000 to 99,000,000 inclusive
			yieldPct: rollDice(r, 10, 6),
		}, nil
	default:
		return &deposit{kind: domain.NONMETALLICS,
			quantity: r.IntN(98_000_001) + 1_000_000, // 1,000,000 to 99,000,000 inclusive
			yieldPct: rollDice(r, 10, 6),
		}, nil
	}
}

func generateTerrestrialDeposits(r *prng.Rand, habitability int) (*deposit, error) {
	roll := r.IntN(100) + 1 // 1...100
	switch {
	case roll == 1:
		d := &deposit{kind: domain.GOLD,
			quantity: r.IntN(900_001) + 100_000, // 100,000 to 1,000,000 inclusive
		}
		if habitability == 0 {
			d.yieldPct = rollDice(r, 1, 3)
		} else {
			d.yieldPct = rollDice(r, 3, 4) - 3
		}
		return d, nil
	case roll <= 15:
		d := &deposit{kind: domain.FUEL,
			quantity: r.IntN(98_000_001) + 1_000_000, // 1,000,000 to 99,000,000 inclusive
		}
		if habitability == 0 {
			d.yieldPct = rollDice(r, 10, 4) - 2
		} else {
			d.yieldPct = rollDice(r, 10, 8)
		}
		return d, nil
	case roll <= 45:
		d := &deposit{kind: domain.METALLICS,
			quantity: r.IntN(98_000_001) + 1_000_000, // 1,000,000 to 99,000,000 inclusive
		}
		if habitability == 0 {
			d.yieldPct = rollDice(r, 10, 6)
		} else {
			d.yieldPct = rollDice(r, 10, 8)
		}
		return d, nil
	default:
		d := &deposit{kind: domain.NONMETALLICS,
			quantity: r.IntN(98_000_001) + 1_000_000, // 1,000,000 to 99,000,000 inclusive
		}
		if habitability == 0 {
			d.yieldPct = rollDice(r, 10, 6)
		} else {
			d.yieldPct = rollDice(r, 10, 8)
		}
		return d, nil
	}
}

// rollDice rolls n dice each with the given number of sides and returns the sum.
func rollDice(r *prng.Rand, n, sides int) int {
	total := 0
	for i := 0; i < n; i++ {
		total += r.IntN(sides) + 1
	}
	return total
}
