package twobuyer

import (
	"testing"

	"github.com/fangyi-zhou/mpst-examples/common"
)

func TestItRuns(t *testing.T) {
	common.RunEndpoints(common.InitStdoutTracer, "TwoBuyer", MakeEndpoints())
}
