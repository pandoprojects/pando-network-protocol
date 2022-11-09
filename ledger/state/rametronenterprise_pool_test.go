package state

import (
	crand "crypto/rand"
	"math/big"
	"testing"

	"github.com/pandoprojects/pando/core"
)

func TestSampleRametronenterpriseWeight(t *testing.T) {
	N := 10000
	weight := 0
	for i := 0; i < N; i++ {
		stake := new(big.Int).Mul(core.MinRametronenterpriseStakeDeposit, big.NewInt(5*100))
		stake.Div(stake, big.NewInt(4))

		totalStake := new(big.Int).Mul(stake, big.NewInt(5))

		weight += sampleRametronenterpriseWeight(crand.Reader, stake, totalStake)
	}

	if float64(weight)/float64(N) > 80+0.1 || float64(weight)/float64(N) < 80-0.1 {
		t.Fail()
	}
}

func BenchmarkRandInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		stake := new(big.Int).Mul(core.MinRametronenterpriseStakeDeposit, big.NewInt(5*100))
		stake.Div(stake, big.NewInt(4))

		totalStake := new(big.Int).Mul(stake, big.NewInt(5))

		sampleRametronenterpriseWeight(crand.Reader, stake, totalStake)
	}
}
