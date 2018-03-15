package server

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"proj2_f5w9a_h6v9a_q7w9a_r8u8_w1c0b/serverpb"

	"github.com/dgraph-io/badger"
)

func (s *Server) Get(ctx context.Context, in *serverpb.GetRequest) (*serverpb.GetResponse, error) {
	var f serverpb.File
	if err := s.db.View(func(txn *badger.Txn) error {
		key := fmt.Sprintf("/document/%s", in.FileId)
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		body, err := item.Value()
		if err != nil {
			return err
		}
		if err := f.Unmarshal(body); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	resp := &serverpb.GetResponse{
		File: &f,
	}

	return resp, nil
}

func (s *Server) Add(ctx context.Context, in *serverpb.AddRequest) (*serverpb.AddResponse, error) {
	b, err := in.File.Marshal()
	if err != nil {
		return nil, err
	}

	data := sha1.Sum(b)
	hash := base64.StdEncoding.EncodeToString(data[:])

	if err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(fmt.Sprintf("/document/%s", hash)), b)
	}); err != nil {
		return nil, err
	}

	resp := &serverpb.AddResponse{
		FileId: hash,
	}

	return resp, nil
}

func (s *Server) GetPeers(ctx context.Context, in *serverpb.GetPeersRequest) (*serverpb.GetPeersResponse, error) {
	var peers []*serverpb.NodeMeta
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, v := range s.mu.peerMeta {
		peers = append(peers, &v)
	}

	resp := &serverpb.GetPeersResponse{
		Peers: peers,
	}

	return resp, nil
}

func (s *Server) AddPeer(ctx context.Context, in *serverpb.AddPeerRequest) (*serverpb.AddPeerResponse, error) {
	resp := &serverpb.AddPeerResponse{}
	return resp, nil
}

func (s *Server) GetReference(ctx context.Context, in *serverpb.GetReferenceRequest) (*serverpb.GetReferenceResponse, error) {
	s.mu.Lock()
	defer s.mu.Lock()
	if reference, ok := s.mu.references[in.GetReferenceId()]; ok {
		resp := &serverpb.GetReferenceResponse{
			Reference: &reference,
		}
		return resp, nil
	}
	// TODO: Do a network lookup for this reference
	resp := &serverpb.GetReferenceResponse{}
	return resp, nil
}

func (s *Server) AddReference(ctx context.Context, in *serverpb.AddReferenceRequest) (*serverpb.AddReferenceResponse, error) {
	resp := &serverpb.AddReferenceResponse{}
	return resp, nil
}
