// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package domain_test

import (
	"testing"

	"github.com/mdhender/ec/internal/domain"
)

func TestOrdersValidate_HappyPath(t *testing.T) {
	tests := []struct {
		name  string
		order interface{ Validate() error }
	}{
		{
			name: "SetUpOrder",
			order: domain.SetUpOrder{
				OrderKind: domain.OrderKindSetUp,
				ColonyID:  1,
				NewName:   "New Colony Alpha",
				NewKind:   domain.Factory,
			},
		},
		{
			name: "BuildChangeOrder",
			order: domain.BuildChangeOrder{
				OrderKind:      domain.OrderKindBuildChange,
				ColonyID:       1,
				FactoryGroupID: 2,
				NewUnitKind:    domain.Mine,
			},
		},
		{
			name: "MiningChangeOrder",
			order: domain.MiningChangeOrder{
				OrderKind:     domain.OrderKindMiningChange,
				ColonyID:      1,
				MiningGroupID: 2,
				DepositID:     3,
			},
		},
		{
			name: "TransferOrder",
			order: domain.TransferOrder{
				OrderKind: domain.OrderKindTransfer,
				SourceID:  1,
				DestID:    2,
				UnitKind:  domain.Food,
				Quantity:  100,
			},
		},
		{
			name: "AssembleOrder",
			order: domain.AssembleOrder{
				OrderKind: domain.OrderKindAssemble,
				ColonyID:  1,
				UnitKind:  domain.Factory,
				Quantity:  5,
			},
		},
		{
			name: "MoveOrder",
			order: domain.MoveOrder{
				OrderKind:   domain.OrderKindMove,
				ShipID:      1,
				Destination: domain.MoveDestination{Raw: "01-02-03/4"},
			},
		},
		{
			name: "DraftOrder",
			order: domain.DraftOrder{
				OrderKind: domain.OrderKindDraft,
				ColonyID:  1,
				PopKind:   domain.UnskilledWorkers,
				Quantity:  50,
			},
		},
		{
			name: "PayOrder",
			order: domain.PayOrder{
				OrderKind: domain.OrderKindPay,
				ColonyID:  1,
				PopKind:   domain.Professionals,
				Wage:      10,
			},
		},
		{
			name: "PayOrder wage zero",
			order: domain.PayOrder{
				OrderKind: domain.OrderKindPay,
				ColonyID:  1,
				PopKind:   domain.Professionals,
				Wage:      0,
			},
		},
		{
			name: "RationOrder",
			order: domain.RationOrder{
				OrderKind:        domain.OrderKindRation,
				ColonyID:         1,
				RationPercentage: 75,
			},
		},
		{
			name: "RationOrder zero percent",
			order: domain.RationOrder{
				OrderKind:        domain.OrderKindRation,
				ColonyID:         1,
				RationPercentage: 0,
			},
		},
		{
			name: "RationOrder 100 percent",
			order: domain.RationOrder{
				OrderKind:        domain.OrderKindRation,
				ColonyID:         1,
				RationPercentage: 100,
			},
		},
		{
			name: "NameOrder planet",
			order: domain.NameOrder{
				OrderKind:  domain.OrderKindName,
				TargetKind: domain.NameTargetPlanet,
				TargetID:   5,
				NewName:    "New Terra",
			},
		},
		{
			name: "NameOrder ship",
			order: domain.NameOrder{
				OrderKind:  domain.OrderKindName,
				TargetKind: domain.NameTargetShip,
				TargetID:   3,
				NewName:    "Voyager I",
			},
		},
		{
			name: "NameOrder colony",
			order: domain.NameOrder{
				OrderKind:  domain.OrderKindName,
				TargetKind: domain.NameTargetColony,
				TargetID:   7,
				NewName:    "Outpost Beta",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.order.Validate(); err != nil {
				t.Errorf("expected nil error, got %v", err)
			}
		})
	}
}

func TestOrdersValidate_InvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		order interface{ Validate() error }
	}{
		// SetUpOrder
		{
			name: "SetUpOrder zero colony ID",
			order: domain.SetUpOrder{
				OrderKind: domain.OrderKindSetUp,
				ColonyID:  0,
				NewName:   "Alpha",
				NewKind:   domain.Factory,
			},
		},
		{
			name: "SetUpOrder negative colony ID",
			order: domain.SetUpOrder{
				OrderKind: domain.OrderKindSetUp,
				ColonyID:  -1,
				NewName:   "Alpha",
				NewKind:   domain.Factory,
			},
		},
		{
			name: "SetUpOrder empty name",
			order: domain.SetUpOrder{
				OrderKind: domain.OrderKindSetUp,
				ColonyID:  1,
				NewName:   "",
				NewKind:   domain.Factory,
			},
		},
		{
			name: "SetUpOrder invalid kind",
			order: domain.SetUpOrder{
				OrderKind: domain.OrderKindSetUp,
				ColonyID:  1,
				NewName:   "Alpha",
				NewKind:   0,
			},
		},
		{
			name: "SetUpOrder wrong order kind",
			order: domain.SetUpOrder{
				OrderKind: domain.OrderKindBuildChange,
				ColonyID:  1,
				NewName:   "Alpha",
				NewKind:   domain.Factory,
			},
		},
		// BuildChangeOrder
		{
			name: "BuildChangeOrder zero colony ID",
			order: domain.BuildChangeOrder{
				OrderKind:      domain.OrderKindBuildChange,
				ColonyID:       0,
				FactoryGroupID: 1,
				NewUnitKind:    domain.Mine,
			},
		},
		{
			name: "BuildChangeOrder zero factory group ID",
			order: domain.BuildChangeOrder{
				OrderKind:      domain.OrderKindBuildChange,
				ColonyID:       1,
				FactoryGroupID: 0,
				NewUnitKind:    domain.Mine,
			},
		},
		{
			name: "BuildChangeOrder invalid unit kind",
			order: domain.BuildChangeOrder{
				OrderKind:      domain.OrderKindBuildChange,
				ColonyID:       1,
				FactoryGroupID: 1,
				NewUnitKind:    0,
			},
		},
		// MiningChangeOrder
		{
			name: "MiningChangeOrder zero colony ID",
			order: domain.MiningChangeOrder{
				OrderKind:     domain.OrderKindMiningChange,
				ColonyID:      0,
				MiningGroupID: 1,
				DepositID:     1,
			},
		},
		{
			name: "MiningChangeOrder zero mining group ID",
			order: domain.MiningChangeOrder{
				OrderKind:     domain.OrderKindMiningChange,
				ColonyID:      1,
				MiningGroupID: 0,
				DepositID:     1,
			},
		},
		{
			name: "MiningChangeOrder zero deposit ID",
			order: domain.MiningChangeOrder{
				OrderKind:     domain.OrderKindMiningChange,
				ColonyID:      1,
				MiningGroupID: 1,
				DepositID:     0,
			},
		},
		// TransferOrder
		{
			name: "TransferOrder zero source ID",
			order: domain.TransferOrder{
				OrderKind: domain.OrderKindTransfer,
				SourceID:  0,
				DestID:    1,
				UnitKind:  domain.Food,
				Quantity:  10,
			},
		},
		{
			name: "TransferOrder zero dest ID",
			order: domain.TransferOrder{
				OrderKind: domain.OrderKindTransfer,
				SourceID:  1,
				DestID:    0,
				UnitKind:  domain.Food,
				Quantity:  10,
			},
		},
		{
			name: "TransferOrder zero quantity",
			order: domain.TransferOrder{
				OrderKind: domain.OrderKindTransfer,
				SourceID:  1,
				DestID:    2,
				UnitKind:  domain.Food,
				Quantity:  0,
			},
		},
		{
			name: "TransferOrder negative quantity",
			order: domain.TransferOrder{
				OrderKind: domain.OrderKindTransfer,
				SourceID:  1,
				DestID:    2,
				UnitKind:  domain.Food,
				Quantity:  -5,
			},
		},
		// AssembleOrder
		{
			name: "AssembleOrder zero colony ID",
			order: domain.AssembleOrder{
				OrderKind: domain.OrderKindAssemble,
				ColonyID:  0,
				UnitKind:  domain.Factory,
				Quantity:  1,
			},
		},
		{
			name: "AssembleOrder zero quantity",
			order: domain.AssembleOrder{
				OrderKind: domain.OrderKindAssemble,
				ColonyID:  1,
				UnitKind:  domain.Factory,
				Quantity:  0,
			},
		},
		// MoveOrder
		{
			name: "MoveOrder zero ship ID",
			order: domain.MoveOrder{
				OrderKind:   domain.OrderKindMove,
				ShipID:      0,
				Destination: domain.MoveDestination{Raw: "01-02-03/4"},
			},
		},
		{
			name: "MoveOrder empty destination",
			order: domain.MoveOrder{
				OrderKind:   domain.OrderKindMove,
				ShipID:      1,
				Destination: domain.MoveDestination{Raw: ""},
			},
		},
		// DraftOrder
		{
			name: "DraftOrder zero colony ID",
			order: domain.DraftOrder{
				OrderKind: domain.OrderKindDraft,
				ColonyID:  0,
				PopKind:   domain.UnskilledWorkers,
				Quantity:  10,
			},
		},
		{
			name: "DraftOrder zero quantity",
			order: domain.DraftOrder{
				OrderKind: domain.OrderKindDraft,
				ColonyID:  1,
				PopKind:   domain.UnskilledWorkers,
				Quantity:  0,
			},
		},
		{
			name: "DraftOrder negative quantity",
			order: domain.DraftOrder{
				OrderKind: domain.OrderKindDraft,
				ColonyID:  1,
				PopKind:   domain.UnskilledWorkers,
				Quantity:  -1,
			},
		},
		// PayOrder
		{
			name: "PayOrder zero colony ID",
			order: domain.PayOrder{
				OrderKind: domain.OrderKindPay,
				ColonyID:  0,
				PopKind:   domain.Professionals,
				Wage:      5,
			},
		},
		{
			name: "PayOrder negative wage",
			order: domain.PayOrder{
				OrderKind: domain.OrderKindPay,
				ColonyID:  1,
				PopKind:   domain.Professionals,
				Wage:      -1,
			},
		},
		// RationOrder
		{
			name: "RationOrder zero colony ID",
			order: domain.RationOrder{
				OrderKind:        domain.OrderKindRation,
				ColonyID:         0,
				RationPercentage: 50,
			},
		},
		{
			name: "RationOrder percentage over 100",
			order: domain.RationOrder{
				OrderKind:        domain.OrderKindRation,
				ColonyID:         1,
				RationPercentage: 101,
			},
		},
		{
			name: "RationOrder percentage negative",
			order: domain.RationOrder{
				OrderKind:        domain.OrderKindRation,
				ColonyID:         1,
				RationPercentage: -1,
			},
		},
		// NameOrder
		{
			name: "NameOrder zero target kind",
			order: domain.NameOrder{
				OrderKind:  domain.OrderKindName,
				TargetKind: 0,
				TargetID:   1,
				NewName:    "New Name",
			},
		},
		{
			name: "NameOrder zero target ID",
			order: domain.NameOrder{
				OrderKind:  domain.OrderKindName,
				TargetKind: domain.NameTargetPlanet,
				TargetID:   0,
				NewName:    "New Name",
			},
		},
		{
			name: "NameOrder empty new name",
			order: domain.NameOrder{
				OrderKind:  domain.OrderKindName,
				TargetKind: domain.NameTargetPlanet,
				TargetID:   1,
				NewName:    "",
			},
		},
		{
			name: "NameOrder negative target ID",
			order: domain.NameOrder{
				OrderKind:  domain.OrderKindName,
				TargetKind: domain.NameTargetShip,
				TargetID:   -3,
				NewName:    "Voyager",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.order.Validate(); err == nil {
				t.Errorf("expected non-nil error, got nil")
			}
		})
	}
}
