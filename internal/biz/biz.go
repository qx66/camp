package biz

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewInstanceUseCase,
	NewMessageUseCase,
	NewInstructUseCase,
	NewChromeDpClientUseCase,
	NewDnsInspectUseCase,
	NewDnsClientInspectUseCase,
	NewHttpInspectClientUseCase,
	NewIcmpClientUseCase,
	NewSocketClientUseCase,
	NewWebSocketUseCase,
)
