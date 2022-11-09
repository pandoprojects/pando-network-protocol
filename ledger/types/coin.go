package types

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/pandoprojects/pando/common"
)

var (
	Zero    *big.Int
	Hundred *big.Int
)

func init() {
	Zero = big.NewInt(0)
	Hundred = big.NewInt(100)
}

type Coins struct {
	PandoWei *big.Int
	PTXWei *big.Int
}

type CoinsJSON struct {
	PandoWei *common.JSONBig `json:"pandowei"`
	PTXWei *common.JSONBig `json:"ptxwei"`
}

func NewCoinsJSON(coin Coins) CoinsJSON {
	return CoinsJSON{
		PandoWei: (*common.JSONBig)(coin.PandoWei),
		PTXWei: (*common.JSONBig)(coin.PTXWei),
	}
}

func (c CoinsJSON) Coins() Coins {
	return Coins{
		PandoWei: (*big.Int)(c.PandoWei),
		PTXWei: (*big.Int)(c.PTXWei),
	}
}

func (c Coins) MarshalJSON() ([]byte, error) {
	return json.Marshal(NewCoinsJSON(c))
}

func (c *Coins) UnmarshalJSON(data []byte) error {
	var a CoinsJSON
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*c = a.Coins()
	return nil
}

// NewCoins is a convenient method for creating small amount of coins.
func NewCoins(pando int64, ptx int64) Coins {
	return Coins{
		PandoWei: big.NewInt(pando),
		PTXWei: big.NewInt(ptx),
	}
}

func (coins Coins) String() string {
	return fmt.Sprintf("%v %v, %v %v", coins.PandoWei, DenomPandoWei, coins.PTXWei, DenomPTXWei)
}

func (coins Coins) IsValid() bool {
	return coins.IsNonnegative()
}

func (coins Coins) NoNil() Coins {
	pando := coins.PandoWei
	if pando == nil {
		pando = big.NewInt(0)
	}
	ptx := coins.PTXWei
	if ptx == nil {
		ptx = big.NewInt(0)
	}

	return Coins{
		PandoWei: pando,
		PTXWei: ptx,
	}
}

// CalculatePercentage function calculates amount of coins for the given the percentage
func (coins Coins) CalculatePercentage(percentage uint) Coins {
	c := coins.NoNil()

	p := big.NewInt(int64(percentage))

	pando := new(big.Int)
	pando.Mul(c.PandoWei, p)
	pando.Div(pando, Hundred)

	ptx := new(big.Int)
	ptx.Mul(c.PTXWei, p)
	ptx.Div(ptx, Hundred)

	return Coins{
		PandoWei: pando,
		PTXWei: ptx,
	}
}

// Currently appends an empty coin ...
func (coinsA Coins) Plus(coinsB Coins) Coins {
	cA := coinsA.NoNil()
	cB := coinsB.NoNil()

	pando := new(big.Int)
	pando.Add(cA.PandoWei, cB.PandoWei)

	ptx := new(big.Int)
	ptx.Add(cA.PTXWei, cB.PTXWei)

	return Coins{
		PandoWei: pando,
		PTXWei: ptx,
	}
}

func (coins Coins) Negative() Coins {
	c := coins.NoNil()

	pando := new(big.Int)
	pando.Neg(c.PandoWei)

	ptx := new(big.Int)
	ptx.Neg(c.PTXWei)

	return Coins{
		PandoWei: pando,
		PTXWei: ptx,
	}
}

func (coinsA Coins) Minus(coinsB Coins) Coins {
	return coinsA.Plus(coinsB.Negative())
}

func (coinsA Coins) IsGTE(coinsB Coins) bool {
	diff := coinsA.Minus(coinsB)
	return diff.IsNonnegative()
}

func (coins Coins) IsZero() bool {
	c := coins.NoNil()
	return c.PandoWei.Cmp(Zero) == 0 && c.PTXWei.Cmp(Zero) == 0
}

func (coinsA Coins) IsEqual(coinsB Coins) bool {
	cA := coinsA.NoNil()
	cB := coinsB.NoNil()
	return cA.PandoWei.Cmp(cB.PandoWei) == 0 && cA.PTXWei.Cmp(cB.PTXWei) == 0
}

func (coins Coins) IsPositive() bool {
	c := coins.NoNil()
	return (c.PandoWei.Cmp(Zero) > 0 && c.PTXWei.Cmp(Zero) >= 0) ||
		(c.PandoWei.Cmp(Zero) >= 0 && c.PTXWei.Cmp(Zero) > 0)
}

func (coins Coins) IsNonnegative() bool {
	c := coins.NoNil()
	return c.PandoWei.Cmp(Zero) >= 0 && c.PTXWei.Cmp(Zero) >= 0
}

// ParseCoinAmount parses a string representation of coin amount.
func ParseCoinAmount(in string) (*big.Int, bool) {
	inWei := false
	if len(in) > 3 && strings.EqualFold("wei", in[len(in)-3:]) {
		inWei = true
		in = in[:len(in)-3]
	}

	f, ok := new(big.Float).SetPrec(1024).SetString(in)
	if !ok || f.Sign() < 0 {
		return nil, false
	}

	if !inWei {
		f = f.Mul(f, new(big.Float).SetPrec(1024).SetUint64(1e18))
	}

	ret, _ := f.Int(nil)

	return ret, true
}
