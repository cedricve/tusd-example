package examples

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/bdragon300/tusgo"
)

func UploadWithRetry(dst *tusgo.UploadStream, src *os.File) error {
	// Set stream and file pointer to be equal to the remote pointer
	// (if we resume the upload that was interrupted earlier)
	if _, err := dst.Sync(); err != nil {
		return err
	}
	if _, err := src.Seek(dst.Tell(), io.SeekStart); err != nil {
		return err
	}

	_, err := io.Copy(dst, src)
	attempts := 10
	for err != nil && attempts > 0 {
		if _, ok := err.(net.Error); !ok && !errors.Is(err, tusgo.ErrChecksumMismatch) {
			return err // Permanent error, no luck
		}
		time.Sleep(5 * time.Second)
		attempts--
		_, err = io.Copy(dst, src) // Try to resume the transfer again
	}
	if attempts == 0 {
		return errors.New("too many attempts to upload the data")
	}
	return nil
}

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
	//f, err := os.Open("../data/video.mp4")
	f, err := os.Open("../data/long-video.mp4")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// For a new file we can call the remote creation
	u := CreateUploadFromFile(f, cl)
	fmt.Println(u)

	// For an existing file
	//u := tusgo.Upload{Location: "http://localhost:8089/files/153c3aa002f905fceb60aff36f43f16a", RemoteSize: 1024 * 1024}

	stream := tusgo.NewUploadStream(cl, u)
	if err = UploadWithRetry(stream, f); err != nil {
		panic(err)
	}
}
