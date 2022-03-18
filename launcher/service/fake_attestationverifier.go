// Package service is a fake implementation of the AttestationVerifier for testing.
package service

import (
	"context"
	"crypto/rand"
	"fmt"

	servpb "github.com/google/go-tpm-tools/launcher/proto/attestation_verifier/v0"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	_ "embed"
)

// FakeToken is generated by fake_tokens/fake_rsa_token.txt.
//go:embed fake_tokens/fake_rsa_token.txt
var FakeToken []byte

// FakeServer implements the AttestationVerifier service methods. The initial
// connection ID produced by the server will be "0", incrementing with every
// subsequent request to GetParams.
type FakeServer struct {
	// conns maps connection IDs to nonces.
	conns map[string][]byte

	// nextConnID represents the next connection ID the server will produce.
	nextConnID int
}

// Check that FakeServer implements servpb.AttestationVerifierServer.
var _ servpb.AttestationVerifierServer = &FakeServer{}

// New constructs a new FakeServer.
func New() FakeServer {
	fs := FakeServer{}
	fs.conns = make(map[string][]byte)
	fs.nextConnID = 0
	return fs
}

// GetParams requests attestation parameters (including nonce and audience).
func (s *FakeServer) GetParams(ctx context.Context, req *servpb.GetParamsRequest) (*servpb.GetParamsResponse, error) {
	nonce := make([]byte, 32)
	rand.Read(nonce)

	connID := fmt.Sprint(s.nextConnID)
	s.nextConnID++

	s.conns[connID] = nonce

	resp := &servpb.GetParamsResponse{
		ConnId:   connID,
		Nonce:    nonce,
		Audience: "https://fake_attestation_verifier/v0/conn_id/" + connID,
	}

	return resp, nil
}

// Verify verifies the attestation and return an OIDC/JWT token.
func (s *FakeServer) Verify(ctx context.Context, req *servpb.VerifyRequest) (*servpb.VerifyResponse, error) {
	if req.GetConnId() == "" {
		return nil, status.Error(codes.InvalidArgument, "VerifyRequest is missing conn_id")
	}

	if _, ok := s.conns[req.GetConnId()]; !ok {
		return nil, status.Error(codes.InvalidArgument, "conn_id was not found")
	}

	if req.GetAttestation() == nil {
		return nil, status.Error(codes.InvalidArgument, "VerifyRequest is missing attestation")
	}

	// TODO(b/210015375): Return a more realistic fake OIDC token with fake signing key and claims.
	resp := &servpb.VerifyResponse{
		ClaimsToken: FakeToken,
	}

	return resp, nil
}
