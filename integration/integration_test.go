package integration

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"proj2_f5w9a_h6v9a_q7w9a_r8u8_w1c0b/server"
	"proj2_f5w9a_h6v9a_q7w9a_r8u8_w1c0b/serverpb"
	"proj2_f5w9a_h6v9a_q7w9a_r8u8_w1c0b/util"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
	color.NoColor = false
}

type cluster struct {
	t     *testing.T
	Nodes []*server.Server
	Dirs  []string
}

func NewTestCluster(t *testing.T, n int) *cluster {
	c := cluster{
		t: t,
	}
	for i := 0; i < n; i++ {
		dir, err := ioutil.TempDir("", "ipfs-cluster-test")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		s, err := server.New(serverpb.NodeConfig{
			Path:     dir,
			MaxPeers: 10,
		})
		if err != nil {
			t.Fatalf("%+v", err)
		}
		c.Nodes = append(c.Nodes, s)
		c.Dirs = append(c.Dirs, dir)

		go func() {
			if err := s.Listen(":0"); err != nil {
				t.Errorf("%+v", err)
			}
		}()
	}

	for _, node := range c.Nodes[1:] {
		var meta serverpb.NodeMeta
		var err error
		util.SucceedsSoon(t, func() error {
			meta, err = node.NodeMeta()
			if err != nil {
				return err
			}
			if len(meta.Addrs) == 0 {
				return errors.Errorf("no address")
			}
			return nil
		})
		if err := c.Nodes[0].AddNode(meta); err != nil {
			t.Fatalf("%+v", err)
		}
	}

	return &c
}

func (c *cluster) Close() {
	for _, s := range c.Nodes {
		if err := s.Close(); err != nil {
			c.t.Errorf("%+v", err)
		}
	}

	for _, dir := range c.Dirs {
		if err := os.RemoveAll(dir); err != nil {
			c.t.Errorf("%+v", err)
		}
	}
}
