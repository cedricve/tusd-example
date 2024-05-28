package examples

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/bdragon300/tusgo"
)

func CreateUploadFromFile(f *os.File, cl *tusgo.Client) *tusgo.Upload {
	finfo, err := f.Stat()
	if err != nil {
		panic(err)
	}

	u := tusgo.Upload{}
	if _, err = cl.CreateUpload(&u, finfo.Size(), false, nil); err != nil {
		panic(err)
	}
	return &u
}
func TestRun(t *testing.T) {
	baseURL, _ := url.Parse("http://localhost:8089/files")
	cl := tusgo.NewClient(http.DefaultClient, baseURL)

	// Open a file we want to upload
	f, err := os.Open("../data/video.mp4")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	u := CreateUploadFromFile(f, cl)

	s := tusgo.NewUploadStream(cl, u)
	// Set stream and file pointers to be equal to the remote pointer
	if _, err = s.Sync(); err != nil {
		panic(err)
	}
	if _, err = f.Seek(s.Tell(), io.SeekStart); err != nil {
		panic(err)
	}

	written, err := io.Copy(s, f)
	if err != nil {
		panic(fmt.Sprintf("Written %d bytes, error: %s, last response: %v", written, err, s.LastResponse))
	}
	fmt.Printf("Written %d bytes\n", written)
}
