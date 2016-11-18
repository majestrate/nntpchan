package thumbnail

import (
	"os/exec"
	"regexp"
)

// thumbnail by executing an external program
type ExecThumbnailer struct {
	// path to executable
	Exec string
	// regular expression that checks for acceptable infiles
	Accept *regexp.Regexp
	// function to generate arguments to use with external program
	// inf and outf are the filenames of the input and output files respectively
	// if this is nil the command will be passed in 2 arguments, infile and outfile
	GenArgs func(inf, outf string) []string
}

func (exe *ExecThumbnailer) CanThumbnail(infpath string) bool {
	re := exe.Accept.Copy()
	return re.MatchString(infpath)
}

func (exe *ExecThumbnailer) Generate(infpath, outfpath string) (err error) {
	// do sanity check
	if exe.CanThumbnail(infpath) {
		var args []string
		if exe.GenArgs == nil {
			args = []string{infpath, outfpath}
		} else {
			args = exe.GenArgs(infpath, outfpath)
		}
		cmd := exec.Command(exe.Exec, args...)
		_, err = cmd.CombinedOutput()
	} else {
		err = ErrCannotThumbanil
	}
	return
}
