package integration

import (
	"proj2_f5w9a_h6v9a_q7w9a_r8u8_w1c0b/serverpb"
	"proj2_f5w9a_h6v9a_q7w9a_r8u8_w1c0b/util"
	"testing"

	"github.com/pkg/errors"
)

func TestSimpleCluster(t *testing.T) {
	ts := NewTestCluster(t, 1)
	defer ts.Close()
}

func TestCluster(t *testing.T) {
	const nodes = 5
	ts := NewTestCluster(t, nodes)
	defer ts.Close()

	for i, node := range ts.Nodes {
		util.SucceedsSoon(t, func() error {
			got := node.NumConnections()
			want := nodes - 1
			if got != want {
				return errors.Errorf("%d. expected %d connections; got %d", i, want, got)
			}
			return nil
		})
	}
}

func TestClusterMaxPeers(t *testing.T) {
	const nodes = 5
	ts := NewTestCluster(t, nodes, func(c *serverpb.NodeConfig) {
		c.MaxPeers = 3
	})
	defer ts.Close()

	for i, node := range ts.Nodes {
		util.SucceedsSoon(t, func() error {
			got := node.NumConnections()
			want := 3
			if got != want {
				return errors.Errorf("%d. expected %d connections; got %d", i, want, got)
			}
			return nil
		})
	}
}
