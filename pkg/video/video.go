package video

import (
	"fmt"
	"github.com/imkira/go-libav/avformat"
)

// 需要安装 libavutil 包依赖 -- No package 'libavutil' found

func GetVideo(videUrl string) error {
	ctx, err := avformat.NewContextForInput()
	if err != nil {
		return err
	}
	
	err = ctx.OpenInput(videUrl, nil, nil)
	if err != nil {
		return err
	}
	
	err = ctx.FindStreamInfo(nil)
	if err != nil {
		return err
	}
	
	for _, stream := range ctx.Streams() {
		
		fmt.Println("Index: ", stream.Index())
		fmt.Println("StartTime: ", stream.StartTime())
		fmt.Println("Duration: ", stream.Duration())
		fmt.Println("AverageFrameRate: ", stream.AverageFrameRate())
	}
	
	return nil
}
