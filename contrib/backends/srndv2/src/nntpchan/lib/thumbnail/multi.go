package thumbnail

import (
	"errors"
)

var ErrNoThumbnailer = errors.New("no thumbnailer found")

type multiThumbnailer struct {
	impls []Thumbnailer
}

// get the frist matching thumbnailer that works with the given file
// if we can't find one return nil
func (mth *multiThumbnailer) getThumbnailer(fpath string) Thumbnailer {
	for _, th := range mth.impls {
		if th.CanThumbnail(fpath) {
			return th
		}
	}
	return nil
}

func (mth *multiThumbnailer) Generate(infpath, outfpath string) (err error) {
	th := mth.getThumbnailer(infpath)
	if th == nil {
		err = ErrNoThumbnailer
	} else {
		err = th.Generate(infpath, outfpath)
	}
	return
}

func (mth *multiThumbnailer) CanThumbnail(infpath string) bool {
	for _, th := range mth.impls {
		if th.CanThumbnail(infpath) {
			return true
		}
	}
	return false
}

func MuxThumbnailers(th ...Thumbnailer) Thumbnailer {
	return &multiThumbnailer{
		impls: th,
	}
}
