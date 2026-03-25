// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package app

import "github.com/mdhender/ec/internal/domain"

// ClusterReader loads a previously-generated cluster.
type ClusterReader interface {
	ReadCluster(path string) (domain.Cluster, error)
}

// ClusterWriter persists a generated cluster.
type ClusterWriter interface {
	WriteCluster(path string, cluster domain.Cluster, overwrite bool) error
}

// GameWriter persists a game file.
type GameWriter interface {
	WriteGame(path string, game *domain.Game, overwrite bool) error
}
