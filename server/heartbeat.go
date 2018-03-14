package server

import (
	"context"
	"proj2_f5w9a_h6v9a_q7w9a_r8u8_w1c0b/serverpb"
)

func (s *Server) HeartBeat(ctx context.Context, req *serverpb.HeartBeatRequest) (*serverpb.HeartBeatResponse, error) {
	return &serverpb.HeartBeatResponse{}, nil
}
