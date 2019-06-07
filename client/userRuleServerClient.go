package client

import (
	"context"
	"google.golang.org/grpc"
	ruleservice "v2ray.com/core/app/router/command"
)

type RuleServerClient struct {
	ruleservice.RuleServerClient
}

func NewRuleServerClient(client *grpc.ClientConn) *RuleServerClient {
	return &RuleServerClient{
		RuleServerClient: ruleservice.NewRuleServerClient(client),
	}
}

func (s *RuleServerClient) AddUserRelyRule(targettag string, emails []string) error {
	_, err := s.AddUserRule(context.Background(), &ruleservice.AddUserRuleRequest{
		TargetTag: targettag,
		Email:     emails,
	})
	return err
}

func (s *RuleServerClient) RemveUserRelayRule(email []string) error {
	_, err := s.RemoveUserRule(context.Background(), &ruleservice.RemoveUserRequest{
		Email: email,
	})
	return err
}

func (s *RuleServerClient) AddUserAttrMachter(targettag string, code string) error {
	_, err := s.AddAttrMachter(context.Background(), &ruleservice.AddAttrMachterRequest{
		TargetTag: targettag,
		Code:      code,
	})
	return err
}

func (s *RuleServerClient) RemveUserAttrMachter(targettag string) error {
	_, err := s.RemoveAttrMachter(context.Background(), &ruleservice.RemoveAttrMachterRequest{
		TargetTag: targettag,
	})
	return err
}
