package twobuyer

import (
	"github.com/fangyi-zhou/mpst-examples/common"
	"testing"
)

func TestItRuns(t *testing.T) {
	common.RunEndpoints(common.InitStdoutTracer, MakeEndpoints())
}
