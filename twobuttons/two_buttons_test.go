package twobuttons

import (
	"testing"

	"github.com/fangyi-zhou/mpst-examples/common"
)

func TestItRuns(t *testing.T) {
	common.RunEndpoints(common.InitStdoutTracer, "TwoButtons", MakeEndpoints())
}

func TestItRunsMailbox(t *testing.T) {
	common.RunEndpoints(common.InitStdoutTracer, "TwoButtons", MakeMailboxEndpoints())
}
