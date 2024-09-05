package video

import "testing"

func TestGetVideo(t *testing.T) {
	videoUrl := "https://startops-package-repo.oss-cn-hangzhou.aliyuncs.com/tmp/1-3.mp4"
	GetVideo(videoUrl)
}
