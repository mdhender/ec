// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app

import (
	"fmt"
	"math/rand/v2"

	"github.com/mdhender/ec/internal/domain"
	"github.com/mdhender/ec/internal/domain/clustergen"
	"github.com/mdhender/prng"
)

// ClusterService orchestrates cluster generation use cases.
type ClusterService struct {
	Writer ClusterStore
}

// CreateCluster generates one cluster and writes it to disk.
func (s *ClusterService) CreateCluster(seed1, seed2 uint64, dataPath string, overwrite bool) (domain.Cluster, error) {
	r := prng.New(rand.NewPCG(seed1, seed2))
	cluster, err := clustergen.GenerateCluster(r)
	if err != nil {
		return domain.Cluster{}, fmt.Errorf("createCluster: %w", err)
	}
	if err := s.Writer.WriteCluster(dataPath, cluster, overwrite); err != nil {
		return domain.Cluster{}, fmt.Errorf("createCluster: %w", err)
	}
	return cluster, nil
}

// TestCluster runs N iterations and returns aggregated stats.
func (s *ClusterService) TestCluster(seed1, seed2 uint64, iterations int) (*clustergen.ClusterStats, error) {
	r := prng.New(rand.NewPCG(seed1, seed2))
	stats := clustergen.NewClusterStats()
	for i := 0; i < iterations; i++ {
		cluster, err := clustergen.GenerateCluster(r)
		if err != nil {
			return nil, fmt.Errorf("testCluster: iteration %d: %w", i, err)
		}
		stats.Collect(cluster)
	}
	return stats, nil
}
