package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/security"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
	"strings"
	"time"
)

type ChromeDpClientUseCase struct {
	logger *zap.Logger
}

// 启动chromeDp，并监听Action消息，执行Action动作

type Action struct {
}

func NewChromeDpClientUseCase(logger *zap.Logger) *ChromeDpClientUseCase {
	return &ChromeDpClientUseCase{
		logger: logger,
	}
}

const (
	defaultUserAgent            = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"
	defaultSecChUa              = "\"Google Chrome\";v=\"123\", \"Not:A-Brand\";v=\"8\", \"Chromium\";v=\"123\""
	defaultUserDataDir          = "/data/chrome"
	defaultCaptureScreenshotDir = "/data/screenshot"
	singlePageDefaultTimeout    = 60 * time.Second
	defaultTimeout              = 120 * 60 * time.Second
)

func getHeaders() map[string]interface{} {
	headerMap := make(map[string]interface{})
	headerMap["Sec-Ch-Ua"] = defaultSecChUa
	headerMap["User-Agent"] = defaultUserAgent
	return headerMap
}

func getChromeOpts(headless bool) []chromedp.ExecAllocatorOption {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("enable-automation", false), // 关闭自动化测试标识符，此时浏览器将不再显示”Chrome正受到自动化测试软件的控制。
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
		chromedp.UserDataDir(defaultUserDataDir),
		chromedp.UserAgent(defaultUserAgent),
	)
	return opts
}

type UrlInspectInfo struct {
	Url             string                   `json:"url,omitempty"`
	RemoteAddr      string                   `json:"remoteAddr,omitempty"`
	RemotePort      int64                    `json:"remotePort,omitempty"`
	Status          int64                    `json:"status,omitempty"`
	Protocol        string                   `json:"protocol,omitempty"`
	Timing          *network.ResourceTiming  `json:"timing,omitempty"`
	FromDiskCache   bool                     `json:"fromDiskCache,omitempty"`
	SecurityState   security.State           `json:"securityState,omitempty"`
	SecurityDetails *network.SecurityDetails `json:"securityDetails,omitempty"`
}

type InspectSinglePageResp struct {
	Url                 string
	HomePageInspect     UrlInspectInfo
	ResourcePageInspect []UrlInspectInfo
}

func (inspectSinglePageResp *InspectSinglePageResp) String() string {
	r, err := json.Marshal(inspectSinglePageResp)
	if err != nil {
		return err.Error()
	}
	
	return string(r)
}

// 使用chromeDp进行单页面检测

func (chromeDpClientUseCase *ChromeDpClientUseCase) InspectSinglePage(ctx context.Context, url string) (InspectSinglePageResp, error) {
	chromeDpClientUseCase.logger.Info("开始执行chromeDp任务")
	var inspectSinglePageResp InspectSinglePageResp
	inspectSinglePageResp.Url = url
	
	opts := getChromeOpts(true)
	ctx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()
	
	// ctx
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	
	// timeout - 不需要timeout，或者设置一个超长timeout，或者根据用户购买时长设置一个timeout
	ctx, cancel = context.WithTimeout(ctx, singlePageDefaultTimeout)
	defer cancel()
	
	//
	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch event := v.(type) {
		case *network.EventRequestWillBeSent:
			//r, _ := event.MarshalJSON()
			//fmt.Println("发送请求数据: ", string(r))
		
		case *network.EventResponseReceived:
			//r, _ := event.MarshalJSON()
			//fmt.Println("接收请求数据: ", string(r))
			if strings.Trim(strings.TrimSpace(event.Response.URL), "/") == url {
				inspectSinglePageResp.HomePageInspect = UrlInspectInfo{
					Url:             event.Response.URL,
					RemoteAddr:      event.Response.RemoteIPAddress,
					RemotePort:      event.Response.RemotePort,
					Status:          event.Response.Status,
					Protocol:        event.Response.Protocol,
					Timing:          event.Response.Timing,
					FromDiskCache:   event.Response.FromDiskCache,
					SecurityState:   event.Response.SecurityState,
					SecurityDetails: event.Response.SecurityDetails,
				}
			} else {
				inspectSinglePageResp.ResourcePageInspect = append(inspectSinglePageResp.ResourcePageInspect, UrlInspectInfo{
					Url:             event.Response.URL,
					RemoteAddr:      event.Response.RemoteIPAddress,
					RemotePort:      event.Response.RemotePort,
					Status:          event.Response.Status,
					Protocol:        event.Response.Protocol,
					Timing:          event.Response.Timing,
					FromDiskCache:   event.Response.FromDiskCache,
					SecurityState:   event.Response.SecurityState,
					SecurityDetails: event.Response.SecurityDetails,
				})
			}
		
		case *network.EventLoadingFinished:
			
		}
	})
	
	tasks := chromedp.Tasks{
		network.Enable(),
		network.SetExtraHTTPHeaders(getHeaders()),
		chromedp.Navigate(url),
	}
	//
	err := chromedp.Run(ctx, tasks)
	if err != nil {
		chromeDpClientUseCase.logger.Error("执行chromeDp任务失败", zap.Error(err))
	}
	
	return inspectSinglePageResp, err
}

func (chromeDpClientUseCase *ChromeDpClientUseCase) newChromeDpClient(ctx context.Context, mainUrl string, actionChannel chan Action, sendMsg chan ClientMessage) {
	
	chromeDpClientUseCase.logger.Info("开始执行chromeDp任务")
	opts := getChromeOpts(false)
	ctx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()
	
	// ctx
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	
	// timeout - 不需要timeout，或者设置一个超长timeout，或者根据用户购买时长设置一个timeout
	ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	
	//
	chromeDpClientUseCase.ListenTarget(ctx)
	
	// task
	tasks := chromedp.Tasks{
		network.Enable(),
		network.SetExtraHTTPHeaders(getHeaders()),
		chromedp.Navigate(mainUrl),
	}
	//
	err := chromedp.Run(ctx, tasks)
	if err != nil {
		chromeDpClientUseCase.logger.Error("执行chromeDp任务失败", zap.Error(err))
		return
	}
	
	// 定时截图
	go chromeDpClientUseCase.CaptureScreenshot(ctx, sendMsg)
	
	// 执行动作
	go chromeDpClientUseCase.execAction(ctx, actionChannel)
	
	select {
	case <-ctx.Done():
		chromeDpClientUseCase.logger.Info("退出chromeDp程序")
		return
	}
}

// 监听 ActionChannel 并执行动作

func (chromeDpClientUseCase *ChromeDpClientUseCase) execAction(ctx context.Context, actionChannel chan Action) {
	// 监听 actionChannel，并执行对应的动作
	for {
		select {
		case act := <-actionChannel:
			chromeDpClientUseCase.logger.Info("action", zap.Any("action", act))
		
		case <-ctx.Done():
			chromeDpClientUseCase.logger.Info("退出监听chromeDp程序")
			return
		}
	}
}

// 每秒截图

func (chromeDpClientUseCase *ChromeDpClientUseCase) CaptureScreenshot(ctx context.Context, sendMsg chan ClientMessage) {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			var img []byte
			
			err := chromedp.Run(ctx, chromedp.CaptureScreenshot(&img))
			if err != nil {
				chromeDpClientUseCase.logger.Error("执行定时截图任务失败", zap.Error(err))
			} else {
				sendMsg <- ClientMessage{
					Type:               ClientChromeDpScreenShot,
					ChromeDpScreenShot: img,
				}
			}
		
		case <-ctx.Done():
			chromeDpClientUseCase.logger.Info("执行定时截图任务退出")
			return
		}
	}
}

func (chromeDpClientUseCase *ChromeDpClientUseCase) SendKey(ctx context.Context, sel interface{}, value string) error {
	return chromedp.Run(ctx, chromedp.SendKeys(sel, value, chromedp.ByQuery))
}

func (chromeDpClientUseCase *ChromeDpClientUseCase) Click(ctx context.Context, sel interface{}) error {
	return chromedp.Run(ctx, chromedp.Click(sel, chromedp.ByQuery))
}

func (chromeDpClientUseCase *ChromeDpClientUseCase) ListenTarget(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch event := v.(type) {
		case *network.EventRequestWillBeSent:
			chromeDpClientUseCase.logger.Info("EventRequestWillBeSent",
				zap.Any("RequestID", event.RequestID),
				zap.Any("URL", event.Request.URL))
		
		case *network.EventLoadingFinished:
			chromeDpClientUseCase.logger.Info("EventLoadingFinished", zap.Any("RequestID", event.RequestID))
		}
	})
}

// 获取访问页面Url
// RemoteAddr、
// Status、
// Protocol 、

// timing、 -
// {"requestTime":276769.720169,"proxyStart":-1,"proxyEnd":-1,"dnsStart":-1,"dnsEnd":-1,"connectStart":-1,
//  "connectEnd":-1,"sslStart":-1,"sslEnd":-1,"workerStart":-1,"workerReady":-1,"workerFetchStart":-1,
//  "workerRespondWithSettled":-1,"sendStart":0.024,"sendEnd":0.024,"pushStart":0,"pushEnd":0,
//  "receiveHeadersStart":0.264,"receiveHeadersEnd":0.288}
// requestTime: 请求开始的时间戳（相对于某个基准时间），它表示浏览器发起此请求的具体时间
// proxyStart, proxyEnd: 代理服务器的起始和结束时间。 如果值为 -1，表示此请求未经过代理服务器，或代理时间不适用。
// dnsStart dnsEnd: DNS 查询开始和结束时间。 如果值为 -1，表示此请求没有进行 DNS 查询，可能是因为该域名的解析结果已缓存。
// connectStart connectEnd: TCP 连接的开始和结束时间，包括与服务器的连接建立时间。 如果值为 -1，表示没有新的 TCP 连接，因为可能复用了已有的连接（如 HTTP/2）
// sslStart sslEnd: SSL/TLS 握手开始和结束时间。 如果值为 -1，表示此请求未使用 SSL/TLS（例如 HTTP 请求）或复用了已有连接的 SSL/TLS 会话。
// workerStart，workerReady，workerFetchStart，workerRespondWithSettled:
// workerStart: Service Worker 开始处理请求的时间。
// workerReady: Service Worker 准备好处理请求的时间。
// workerFetchStart: Service Worker 开始 fetch 请求的时间。
// workerRespondWithSettled: Service Worker 响应处理完成的时间
// 如果值为 -1，表示请求没有经过 Service Worker。

// sendStart 和 sendEnd: 浏览器开始发送请求的时间 (sendStart) 和请求发送完成的时间 (sendEnd)。 在你的数据中，sendStart 和 sendEnd 都是 0.024，表示请求几乎瞬间就发送完毕。
// pushStart 和 pushEnd: HTTP/2 Push 的开始和结束时间。 如果值为 0，表示此请求没有使用 HTTP/2 Push。
// receiveHeadersStart 和 receiveHeadersEnd:
//浏览器开始接收服务器返回的响应头部数据的时间 (receiveHeadersStart) 和接收完毕的时间 (receiveHeadersEnd)。
//在你的数据中，receiveHeadersStart 是 0.264，receiveHeadersEnd 是 0.288，表示从浏览器开始接收响应头到结束的时间是 0.024 秒。

// fromDiskCache、
// securityState、
// securityDetails - protocol 、cipher 、subjectName 、issuer 、validFrom 、validTo

func (chromeDpClientUseCase *ChromeDpClientUseCase) InspectTarget(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch event := v.(type) {
		case *network.EventRequestWillBeSent:
			//r, _ := event.MarshalJSON()
			//fmt.Println("发送请求数据: ", string(r))
		
		case *network.EventResponseReceived:
			r, _ := event.MarshalJSON()
			fmt.Println("接收请求数据: ", string(r))
		
		case *network.EventLoadingFinished:
			
		}
	})
}

// 获取访问页面资源信息
