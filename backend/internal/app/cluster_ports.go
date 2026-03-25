// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app

import "github.com/mdhender/ec/internal/domain"

// ClusterStore reads and writes cluster files.
type ClusterStore interface {
	ReadCluster(dataPath string) (domain.Cluster, error)
	WriteCluster(dataPath string, cluster domain.Cluster, overwrite bool) error
}
