package dtp

import (
	"testing"

	dtp "github.com/WhilecodingDoLearn/dtp/protocol/types"
	"github.com/stretchr/testify/assert"
)

func TestHandle(t *testing.T) {

	tests := []struct {
		name       string
		pck        dtp.Package
		expSize    int
		expSession bool
		expState   dtp.State
	}{{
		"Start Connection",
		dtp.Package{Sid: 1234, Msg: dtp.REQ},
		1,
		true,
		dtp.REQ,
	}, {
		"Send Another Start Connection",
		dtp.Package{Sid: 1234, Msg: dtp.REQ},
		1,
		true,
		dtp.OPN,
	},
	}

	sh := SessionHandler{sessionCache: make(map[int]*Session)}
	for _, subtest := range tests {
		t.Run(subtest.name, func(t *testing.T) {
			handle(subtest.pck, &sh)
			assert.Equal(t, subtest.expSize, sh.Size(), "expection of session count was wrong")
			session, ok := sh.GetSession(1234)
			assert.Equal(t, subtest.expSession, ok, "expection of session  was wrong")
			assert.Equal(t, subtest.expState, session.state, "wrong session state")

		})
	}

}
