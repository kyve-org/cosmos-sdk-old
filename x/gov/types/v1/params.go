package v1

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Default period for deposits & voting
const (
	DefaultPeriod          time.Duration = time.Hour * 24 * 2 // 2 days
	DefaultExpeditedPeriod time.Duration = time.Hour * 24     // 1 day
)

// Default governance params
var (
	DefaultMinDepositTokens          = sdk.NewInt(10000000)
	DefaultMinExpeditedDepositTokens = sdk.NewInt(10000000 * 5)
	DefaultMinDepositPercentage      = sdk.ZeroDec()
	DefaultQuorum                    = sdk.NewDecWithPrec(334, 3)
	DefaultThreshold                 = sdk.NewDecWithPrec(5, 1)
	DefaultExpeditedThreshold        = sdk.NewDecWithPrec(667, 3)
	DefaultVetoThreshold             = sdk.NewDecWithPrec(334, 3)
)

// Parameter store key
var (
	ParamStoreKeyDepositParams = []byte("depositparams")
	ParamStoreKeyVotingParams  = []byte("votingparams")
	ParamStoreKeyTallyParams   = []byte("tallyparams")
)

// ParamKeyTable - Key declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable(
		paramtypes.NewParamSetPair(ParamStoreKeyDepositParams, DepositParams{}, validateDepositParams),
		paramtypes.NewParamSetPair(ParamStoreKeyVotingParams, VotingParams{}, validateVotingParams),
		paramtypes.NewParamSetPair(ParamStoreKeyTallyParams, TallyParams{}, validateTallyParams),
	)
}

// NewDepositParams creates a new DepositParams object
func NewDepositParams(minDeposit sdk.Coins, maxDepositPeriod time.Duration, minExpeditedDeposit sdk.Coins, minDepositPercentage sdk.Dec) DepositParams {
	return DepositParams{
		MinDeposit:           minDeposit,
		MaxDepositPeriod:     &maxDepositPeriod,
		MinExpeditedDeposit:  minExpeditedDeposit,
		MinDepositPercentage: minDepositPercentage.String(),
	}
}

// DefaultDepositParams default parameters for deposits
func DefaultDepositParams() DepositParams {
	return NewDepositParams(
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, DefaultMinDepositTokens)),
		DefaultPeriod,
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, DefaultMinExpeditedDepositTokens)),
		DefaultMinDepositPercentage,
	)
}

// GetMinDeposit returns minimum deposit based on whether isExpedited is requested.
func (dp DepositParams) GetMinDeposit(isExpedited bool) sdk.Coins {
	if isExpedited {
		return dp.MinExpeditedDeposit
	}
	return dp.MinDeposit
}

// GetAdjustedDeposit returns the required minimum deposit needed when submitting a proposal.
func (dp DepositParams) GetAdjustedDeposit(isExpedited bool) sdk.Coins {
	minDeposit := dp.GetMinDeposit(isExpedited)
	minDepositPercentage, _ := sdk.NewDecFromStr(dp.MinDepositPercentage)

	adjustedMinDeposit := sdk.NewCoins()
	for _, coin := range minDeposit {
		amount := sdk.NewDecFromInt(coin.Amount).Mul(minDepositPercentage).RoundInt()
		adjustedMinDeposit = adjustedMinDeposit.Add(sdk.NewCoin(coin.Denom, amount))
	}

	return adjustedMinDeposit
}

// Equal checks equality of DepositParams
func (dp DepositParams) Equal(dp2 DepositParams) bool {
	return sdk.Coins(dp.MinDeposit).IsEqual(dp2.MinDeposit) && dp.MaxDepositPeriod == dp2.MaxDepositPeriod && sdk.Coins(dp.MinExpeditedDeposit).IsEqual(dp2.MinExpeditedDeposit) && dp.MinDepositPercentage == dp2.MinDepositPercentage
}

func validateDepositParams(i interface{}) error {
	v, ok := i.(DepositParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if !sdk.Coins(v.MinDeposit).IsValid() {
		return fmt.Errorf("invalid minimum deposit: %s", v.MinDeposit)
	}

	if v.MaxDepositPeriod == nil || v.MaxDepositPeriod.Seconds() <= 0 {
		return fmt.Errorf("maximum deposit period must be positive: %d", v.MaxDepositPeriod)
	}

	if !sdk.Coins(v.MinExpeditedDeposit).IsValid() {
		return fmt.Errorf("invalid minimum expedited deposit: %s", v.MinExpeditedDeposit)
	}
	if sdk.Coins(v.MinExpeditedDeposit).IsAllLTE(v.MinDeposit) {
		return fmt.Errorf("minimum expedited deposit %s, must be greater than regular deposit %s", v.MinExpeditedDeposit, v.MinDeposit)
	}

	minDepositPercentage, err := sdk.NewDecFromStr(v.MinDepositPercentage)
	if err != nil {
		return fmt.Errorf("invalid min deposit percentage string: %w", err)
	}
	if minDepositPercentage.IsNegative() {
		return fmt.Errorf("min deposit percentage cannot be negative: %s", minDepositPercentage)
	}
	if minDepositPercentage.GT(sdk.OneDec()) {
		return fmt.Errorf("min deposit percentage too large: %s", v)
	}

	return nil
}

// NewTallyParams creates a new TallyParams object
func NewTallyParams(quorum, threshold, expeditedThreshold, vetoThreshold sdk.Dec) TallyParams {
	return TallyParams{
		Quorum:             quorum.String(),
		Threshold:          threshold.String(),
		ExpeditedThreshold: expeditedThreshold.String(),
		VetoThreshold:      vetoThreshold.String(),
	}
}

// DefaultTallyParams default parameters for tallying
func DefaultTallyParams() TallyParams {
	return NewTallyParams(DefaultQuorum, DefaultThreshold, DefaultExpeditedThreshold, DefaultVetoThreshold)
}

// GetThreshold returns threshold based on the value isExpedited
func (tp TallyParams) GetThreshold(isExpedited bool) sdk.Dec {
	if isExpedited {
		expeditedThreshold, _ := sdk.NewDecFromStr(tp.ExpeditedThreshold)
		return expeditedThreshold
	}

	threshold, _ := sdk.NewDecFromStr(tp.Threshold)
	return threshold
}

// Equal checks equality of TallyParams
func (tp TallyParams) Equal(other TallyParams) bool {
	return tp.Quorum == other.Quorum && tp.Threshold == other.Threshold && tp.ExpeditedThreshold == other.ExpeditedThreshold && tp.VetoThreshold == other.VetoThreshold
}

func validateTallyParams(i interface{}) error {
	v, ok := i.(TallyParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	quorum, err := sdk.NewDecFromStr(v.Quorum)
	if err != nil {
		return fmt.Errorf("invalid quorum string: %w", err)
	}
	if quorum.IsNegative() {
		return fmt.Errorf("quorom cannot be negative: %s", quorum)
	}
	if quorum.GT(sdk.OneDec()) {
		return fmt.Errorf("quorom too large: %s", v)
	}

	threshold, err := sdk.NewDecFromStr(v.Threshold)
	if err != nil {
		return fmt.Errorf("invalid threshold string: %w", err)
	}
	if !threshold.IsPositive() {
		return fmt.Errorf("vote threshold must be positive: %s", threshold)
	}
	if threshold.GT(sdk.OneDec()) {
		return fmt.Errorf("vote threshold too large: %s", v)
	}

	expeditedThreshold, err := sdk.NewDecFromStr(v.ExpeditedThreshold)
	if err != nil {
		return fmt.Errorf("invalid expedited threshold string: %w", err)
	}
	if !expeditedThreshold.IsPositive() {
		return fmt.Errorf("expedited vote threshold must be positive: %s", expeditedThreshold)
	}
	if expeditedThreshold.GT(sdk.OneDec()) {
		return fmt.Errorf("expedited vote threshold too large: %s", v)
	}
	if expeditedThreshold.LTE(threshold) {
		return fmt.Errorf("expedited vote threshold %s, must be greater than the regular vote threshold %s", expeditedThreshold, threshold)
	}

	vetoThreshold, err := sdk.NewDecFromStr(v.VetoThreshold)
	if err != nil {
		return fmt.Errorf("invalid vetoThreshold string: %w", err)
	}
	if !vetoThreshold.IsPositive() {
		return fmt.Errorf("veto threshold must be positive: %s", vetoThreshold)
	}
	if vetoThreshold.GT(sdk.OneDec()) {
		return fmt.Errorf("veto threshold too large: %s", v)
	}

	return nil
}

// NewVotingParams creates a new VotingParams object
func NewVotingParams(votingPeriod time.Duration, expeditedPeriod time.Duration) VotingParams {
	return VotingParams{
		VotingPeriod:          &votingPeriod,
		ExpeditedVotingPeriod: &expeditedPeriod,
	}
}

// DefaultVotingParams default parameters for voting
func DefaultVotingParams() VotingParams {
	return NewVotingParams(DefaultPeriod, DefaultExpeditedPeriod)
}

// GetVotingPeriod returns voting period based on whether isExpedited is requested.
func (vp VotingParams) GetVotingPeriod(isExpedited bool) time.Duration {
	if isExpedited {
		return *vp.ExpeditedVotingPeriod
	}
	return *vp.VotingPeriod
}

// Equal checks equality of TallyParams
func (vp VotingParams) Equal(other VotingParams) bool {
	return vp.VotingPeriod == other.VotingPeriod && vp.ExpeditedVotingPeriod == other.ExpeditedVotingPeriod
}

func validateVotingParams(i interface{}) error {
	v, ok := i.(VotingParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.VotingPeriod == nil {
		return errors.New("voting period must not be nil")
	}
	if v.VotingPeriod.Seconds() <= 0 {
		return fmt.Errorf("voting period must be positive: %s", v.VotingPeriod)
	}

	if v.ExpeditedVotingPeriod == nil {
		return errors.New("expedited voting period must not be nil")
	}
	if v.ExpeditedVotingPeriod.Seconds() <= 0 {
		return fmt.Errorf("expedited voting period must be positive: %s", v.ExpeditedVotingPeriod)
	}

	if v.ExpeditedVotingPeriod.Seconds() >= v.VotingPeriod.Seconds() {
		return fmt.Errorf("expedited voting period %s must be strictly less that the regular voting period %s", v.ExpeditedVotingPeriod, v.VotingPeriod)
	}

	return nil
}

// Params returns all of the governance params
type Params struct {
	VotingParams  VotingParams  `json:"voting_params" yaml:"voting_params"`
	TallyParams   TallyParams   `json:"tally_params" yaml:"tally_params"`
	DepositParams DepositParams `json:"deposit_params" yaml:"deposit_params"`
}

func (gp Params) String() string {
	return gp.VotingParams.String() + "\n" +
		gp.TallyParams.String() + "\n" + gp.DepositParams.String()
}

// NewParams creates a new gov Params instance
func NewParams(vp VotingParams, tp TallyParams, dp DepositParams) Params {
	return Params{
		VotingParams:  vp,
		DepositParams: dp,
		TallyParams:   tp,
	}
}

// DefaultParams default governance params
func DefaultParams() Params {
	return NewParams(DefaultVotingParams(), DefaultTallyParams(), DefaultDepositParams())
}
