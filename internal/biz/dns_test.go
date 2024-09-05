package biz

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLookupHost(t *testing.T) {
	dnsClientInspectUseCase := &DnsClientInspectUseCase{}
	fqdn := "google.com"
	r, err := dnsClientInspectUseCase.LookupHost(context.Background(), fqdn)
	require.NoError(t, err, fmt.Sprintf("解析%s失败", fqdn))
	fmt.Println("r: ", r)
}

func TestQuery(t *testing.T) {
	dnsClientInspectUseCase := &DnsClientInspectUseCase{}
	fqdn := "google.com"
	ns := "1.1.1.1:53"
	r, rtt, err := dnsClientInspectUseCase.Query(fqdn, ns)
	require.NoError(t, err, fmt.Sprintf("通过%s解析%s失败", ns, fqdn))
	
	fmt.Println("rtt: ", rtt)
	fmt.Println("result: ", r)
}

func TestLookupHost1(t *testing.T) {
	dnsClientInspectUseCase := &DnsClientInspectUseCase{}
	fqdn := "whoami.akamai.net"
	r, err := dnsClientInspectUseCase.LookupHost(context.Background(), fqdn)
	require.NoError(t, err, "获取本地dns失败")
	fmt.Println("r: ", r)
}

func TestQuery1(t *testing.T) {
	dnsClientInspectUseCase := &DnsClientInspectUseCase{}
	
	fqdn := "whoami.akamai.net"
	ns := "1.1.1.1:53"
	r, rtt, err := dnsClientInspectUseCase.Query(fqdn, ns)
	require.NoError(t, err, fmt.Sprintf("通过%s获取local dns失败", ns))
	
	fmt.Println("rtt: ", rtt)
	fmt.Println("result: ", r)
}
