package thumbnail

import (
	"testing"
)

func doTestThumb(t *testing.T, th Thumbnailer, allowed, disallowed []string) {
	for _, f := range allowed {
		if !th.CanThumbnail(f) {
			t.Logf("cannot thumbnail expected file: %s", f)
			t.Fail()
		}
	}

	for _, f := range disallowed {
		if th.CanThumbnail(f) {
			t.Logf("can thumbnail wrong file: %s", f)
			t.Fail()
		}
	}

}

var _image = []string{"asd.gif", "asd.jpeg", "asd.jpg", "asd.png", "asd.webp"}
var _video = []string{"asd.mkv", "asd.mov", "asd.mp4", "asd.m4v", "asd.ogv", "asd.avi", "asd.mpg", "asd.webm"}
var _sound = []string{"asd.flac", "asd.mp3", "asd.mp2", "asd.wav", "asd.ogg", "asd.opus", "asd.m4a"}
var _misc = []string{"asd.txt", "asd.swf"}
var _garbage = []string{"asd", "asd.asd", "asd.asd.asd.asd", "asd.benis"}

func TestCanThumbnailImage(t *testing.T) {
	th := ImageMagickThumbnailer("", nil)
	var allowed []string
	var disallowed []string

	allowed = append(allowed, _image...)
	disallowed = append(disallowed, _video...)
	disallowed = append(disallowed, _sound...)
	disallowed = append(disallowed, _misc...)
	disallowed = append(disallowed, _garbage...)

	doTestThumb(t, th, allowed, disallowed)
}

func TestCanThumbnailVideo(t *testing.T) {
	th := FFMpegThumbnailer("", nil)
	var allowed []string
	var disallowed []string

	allowed = append(allowed, _video...)
	disallowed = append(disallowed, _image...)
	disallowed = append(disallowed, _sound...)
	disallowed = append(disallowed, _misc...)
	disallowed = append(disallowed, _garbage...)

	doTestThumb(t, th, allowed, disallowed)
}
