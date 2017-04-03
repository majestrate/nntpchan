package srnd

import "testing"

func TestGenFeedsConfig(t *testing.T) {

	err := GenFeedsConfig()
	// Generate default feeds.ini
	if err != nil {

		t.Error("Cannot generate feeds.ini", err)

	}

}

// func (self lineWriter) Write(data []byte) (n int, err error) {

//func OpenFileWriter(fname string) (io.WriteCloser, error) {

func TestOpenFileWriter(t *testing.T) {

	_, err := OpenFileWriter("file.txt")
	// Generate default feeds.ini
	if err != nil {

		t.Error("Cant open file writer.", err)

	}

}
