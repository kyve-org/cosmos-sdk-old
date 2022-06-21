package types

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestProposalStatus_Format(t *testing.T) {
	statusDepositPeriod, _ := ProposalStatusFromString("PROPOSAL_STATUS_DEPOSIT_PERIOD")
	tests := []struct {
		pt                   ProposalStatus
		sprintFArgs          string
		expectedStringOutput string
	}{
		{statusDepositPeriod, "%s", "PROPOSAL_STATUS_DEPOSIT_PERIOD"},
		{statusDepositPeriod, "%v", "1"},
	}
	for _, tt := range tests {
		got := fmt.Sprintf(tt.sprintFArgs, tt.pt)
		require.Equal(t, tt.expectedStringOutput, got)
	}
}

func TestProposalGetMinDepositFromParams(t *testing.T) {
	testcases := []struct {
		isExpedited        bool
		expectedMinDeposit sdk.Int
	}{
		{
			isExpedited:        true,
			expectedMinDeposit: DefaultMinExpeditedDepositTokens,
		},
		{
			isExpedited:        false,
			expectedMinDeposit: DefaultMinDepositTokens,
		},
	}

	for _, tc := range testcases {
		testProposal := NewTextProposal("test", "description")

		proposal, err := NewProposal(testProposal, 1, time.Now(), time.Now(), tc.isExpedited)
		require.NoError(t, err)

		actualMinDeposit := proposal.GetMinDepositFromParams(DefaultDepositParams())

		require.Equal(t, 1, len(actualMinDeposit))
		require.Equal(t, sdk.DefaultBondDenom, actualMinDeposit[0].Denom)
		require.Equal(t, tc.expectedMinDeposit, actualMinDeposit[0].Amount)
	}
}
