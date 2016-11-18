package thumbnail

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var ErrCannotThumbanil = errors.New("cannot thumbnail file")

// a generator of thumbnails
type Thumbnailer interface {
	// generate thumbnail of attachment
	//
	// infpath: absolute filepath to attachment
	//
	// outfpath: absolute filepath to thumbnail
	//
	// return error if the thumbnailing fails
	Generate(infpath, outfpath string) error

	// can we generate a thumbnail for this file?
	CanThumbnail(infpath string) bool
}

// thumbnail configuration
type Config struct {
	// width of thumbnails
	ThumbW int
	// height of thumbnails
	ThumbH int
	// only generate jpg thumbnails
	JpegOnly bool
}

var defaultCfg = &Config{
	ThumbW:   300,
	ThumbH:   200,
	JpegOnly: true,
}

// create an imagemagick thumbnailer
func ImageMagickThumbnailer(convertPath string, conf *Config) Thumbnailer {
	if conf == nil {
		conf = defaultCfg
	}
	return &ExecThumbnailer{
		Exec:   convertPath,
		Accept: regexp.MustCompilePOSIX(`\.(png|jpg|jpeg|gif|webp)$`),
		GenArgs: func(inf, outf string) []string {
			if strings.HasSuffix(inf, ".gif") {
				inf += "[0]"
			}
			if conf.JpegOnly {
				outf += ".jpeg"
			}
			return []string{"-thumbnail", fmt.Sprintf("%d", conf.ThumbW), inf, outf}
		},
	}
}

// generate a thumbnailer that uses ffmpeg
func FFMpegThumbnailer(ffmpegPath string, conf *Config) Thumbnailer {
	if conf == nil {
		conf = defaultCfg
	}
	return &ExecThumbnailer{
		Exec:   ffmpegPath,
		Accept: regexp.MustCompilePOSIX(`\.(mkv|mp4|avi|webm|ogv|mov|m4v|mpg)$`),
		GenArgs: func(inf, outf string) []string {
			outf += ".jpeg"
			return []string{"-i", inf, "-vf", fmt.Sprintf("scale=%d:%d", conf.ThumbW, conf.ThumbH), "-vframes", "1", outf}
		},
	}
}
