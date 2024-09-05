package biz

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIcmp(t *testing.T) {
	icmpClientUseCase := IcmpClientUseCase{}
	r, err := icmpClientUseCase.Icmp("baidu.com", 4)
	require.NoError(t, err, "执行Icmp失败")
	fmt.Println(r)
}
