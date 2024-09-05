package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ndns "github.com/miekg/dns"
	"go.uber.org/zap"
	"net"
)

type DnsInspectUseCase struct {
	logger *zap.Logger
}

type DnsClientInspectUseCase struct {
	logger *zap.Logger
}

// 建议直接使用shell command - 目前未完成

func NewDnsInspectUseCase(logger *zap.Logger) *DnsInspectUseCase {
	return &DnsInspectUseCase{
		logger: logger,
	}
}

func NewDnsClientInspectUseCase(logger *zap.Logger) *DnsClientInspectUseCase {
	return &DnsClientInspectUseCase{
		logger: logger,
	}
}

type DnsInspectReply struct {
	FQDN  string `json:"fqdn,omitempty"`
	A     string `json:"a,omitempty"`
	Cname string `json:"cname,omitempty"`
}

func (dnsInspectReply *DnsInspectReply) String() string {
	b, err := json.Marshal(dnsInspectReply)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (dnsClientInspectUseCase *DnsClientInspectUseCase) LookupHost(ctx context.Context, fqdn string) ([]string, error) {
	resolver := net.Resolver{}
	return resolver.LookupHost(ctx, fqdn)
}

// return result, rtt, error

func (dnsClientInspectUseCase *DnsClientInspectUseCase) Query(fqdn, ns string) (string, float64, error) {
	//
	m := new(ndns.Msg)
	m.SetQuestion(ndns.Fqdn(fqdn), ndns.TypeA)
	m.RecursionDesired = true
	
	//
	cli := new(ndns.Client)
	r, rtt, err := cli.Exchange(m, ns)
	rt := rtt.Seconds()
	
	if err != nil {
		return "", rt, err
	}
	
	if r.Rcode != ndns.RcodeSuccess {
		return "", rt, errors.New(fmt.Sprintf("解析失败, Rcode: %d", r.Rcode))
	}
	
	return r.String(), rt, nil
}
