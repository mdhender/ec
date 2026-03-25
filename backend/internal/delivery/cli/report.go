// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package cli

import (
	"fmt"
	"io"

	"github.com/mdhender/ec/internal/domain"
	"github.com/mdhender/ec/internal/domain/clustergen"
)

// WriteClusterReport generates and writes a single-run cluster report.
func WriteClusterReport(w io.Writer, cluster domain.Cluster) {
	cs := clustergen.NewClusterStats()
	cs.Collect(cluster)
	WriteStatsReport(w, cs, 1)
}

// WriteStatsReport writes a formatted cluster statistics report.
func WriteStatsReport(w io.Writer, cs *clustergen.ClusterStats, divisor int) {
	planetKinds := []domain.PlanetKind{domain.Terrestrial, domain.AsteroidBelt, domain.GasGiant}
	resourceKinds := []domain.NaturalResource{domain.GOLD, domain.FUEL, domain.METALLICS, domain.NONMETALLICS}

	d := float64(divisor)
	numSystems := float64(cs.NumSystems) / d
	numStars := float64(cs.NumStars) / d
	totalPlanets := float64(cs.TotalPlanets) / d

	totalDeposits := 0
	for _, s := range cs.Overall {
		totalDeposits += s.Count
	}

	_, _ = fmt.Fprintln(w, "=== Cluster Report ===")
	if divisor > 1 {
		_, _ = fmt.Fprintf(w, "  (averaged over %d iterations)\n", divisor)
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "  Systems: %.1f\n", numSystems)
	_, _ = fmt.Fprintf(w, "  Stars:   %.1f\n", numStars)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "  Planets:")
	_, _ = fmt.Fprintf(w, "    %-14s %8s %10s %9s %10s\n", "Type", "Count", "Habitable", "Avg Hab", "Per Star")
	totalHabitable := 0
	totalHabSum := 0
	for _, pk := range planetKinds {
		hab := cs.HabitableByKind[pk]
		totalHabitable += hab
		totalHabSum += cs.TotalHabByKind[pk]
		avgHab := "-"
		if hab > 0 {
			avgHab = fmt.Sprintf("%.1f", float64(cs.TotalHabByKind[pk])/float64(hab))
		}
		_, _ = fmt.Fprintf(w, "    %-14s %8.1f %10.1f %9s %10.1f\n", pk, float64(cs.PlanetsByKind[pk])/d, float64(hab)/d, avgHab, float64(cs.PlanetsByKind[pk])/float64(cs.NumStars))
	}
	totalAvgHab := "-"
	if totalHabitable > 0 {
		totalAvgHab = fmt.Sprintf("%.1f", float64(totalHabSum)/float64(totalHabitable))
	}
	_, _ = fmt.Fprintf(w, "    %-14s %8.1f %10.1f %9s %10.1f\n", "Total", totalPlanets, float64(totalHabitable)/d, totalAvgHab, totalPlanets/numStars)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "  Deposits (overall):")
	_, _ = fmt.Fprintf(w, "    %-14s %8s %16s %16s %10s\n", "Resource", "Count", "Total Qty", "Avg Qty", "Avg Yield")
	for _, rk := range resourceKinds {
		s := cs.Overall[rk]
		if s.Count == 0 {
			_, _ = fmt.Fprintf(w, "    %-14s %8s %16s %16s %10s\n", rk, "-", "-", "-", "-")
		} else {
			_, _ = fmt.Fprintf(w, "    %-14s %8.1f %16s %16s %9.1f%%\n", rk, float64(s.Count)/d, commaFmtInt64(s.TotalQty/int64(divisor)), commaFmtInt64(s.TotalQty/int64(s.Count)), float64(s.TotalPct)/float64(s.Count))
		}
	}
	_, _ = fmt.Fprintf(w, "    %-14s %8.1f\n", "Total", float64(totalDeposits)/d)

	for _, pk := range planetKinds {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintf(w, "  Deposits on %s:\n", pk)
		_, _ = fmt.Fprintf(w, "    %-14s %8s %16s %16s %10s\n", "Resource", "Count", "Total Qty", "Avg Qty", "Avg Yield")
		pkTotal := 0
		for _, rk := range resourceKinds {
			s := cs.ByPlanetKind[pk][rk]
			pkTotal += s.Count
			if s.Count == 0 {
				_, _ = fmt.Fprintf(w, "    %-14s %8s %16s %16s %10s\n", rk, "-", "-", "-", "-")
			} else {
				_, _ = fmt.Fprintf(w, "    %-14s %8.1f %16s %16s %9.1f%%\n", rk, float64(s.Count)/d, commaFmtInt64(s.TotalQty/int64(divisor)), commaFmtInt64(s.TotalQty/int64(s.Count)), float64(s.TotalPct)/float64(s.Count))
			}
		}
		_, _ = fmt.Fprintf(w, "    %-14s %8.1f\n", "Total", float64(pkTotal)/d)
	}
}

func commaFmtInt64(n int64) string {
	if n < 0 {
		return "-" + commaFmtInt64(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
