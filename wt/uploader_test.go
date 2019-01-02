package wt

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestUploaderService_upload(t *testing.T) {
	client, mux, srvURL, teardown := setup()
	defer teardown()

	tests := []struct {
		multipartID string
		partNumbers int
		chunkSize   int
		fileID      string
		bot         boardOrTransfer
	}{
		{"", 2, 1024, "1", &Transfer{ID: String("1")}},
		{"99", 2, 256, "2", &Board{ID: String("2")}},
	}

	for _, tt := range tests {
		for i := 1; i <= tt.partNumbers; i++ {
			func(partNum int) {
				var path string
				var s3prefix string
				switch tt.bot.(type) {
				case *Transfer:
					path = fmt.Sprintf("/transfers/%v/files/%v/upload-url/%v", tt.bot.GetID(), tt.fileID, partNum)
					s3prefix = "transfers"
				case *Board:
					path = fmt.Sprintf("/boards/%v/files/%v/upload-url/%v/%v", tt.bot.GetID(), tt.fileID, partNum, tt.multipartID)
					s3prefix = "boards"
				}
				mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintf(w, `{"success": true, "url": "%s"}`, fmt.Sprintf("%v/%v/p/%v", srvURL, s3prefix, partNum))
				})
				// Fake S3 response
				mux.HandleFunc(fmt.Sprintf("/%v/p/%v", s3prefix, partNum), func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
				})
			}(i)
		}

		// File or item agnostic
		file := &File{
			ID:   &tt.fileID,
			Name: String("pony.txt"),
			Multipart: &Multipart{
				ID:          String("99"),
				PartNumbers: Int64(int64(tt.partNumbers)),
				ChunkSize:   Int64(int64(tt.chunkSize)),
			},
		}

		data := make([]byte, tt.partNumbers*tt.chunkSize)
		for i := range data {
			data[i] = 'x'
		}
		buf := NewBuffer("pony.txt", data)
		ft := newFileTransfer(buf, file)
		err := client.uploader.upload(context.Background(), tt.bot, ft)
		if err != nil {
			t.Errorf("upload returned an error: %+v", err)
		}
	}
}

func TestUploaderService_getUploadURL_Transfers(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	ts := &Transfer{ID: String("1")}
	board := &Board{ID: String("2")}

	tests := []struct {
		partNum int64
		fid     string
		mid     string
		bot     boardOrTransfer
		uurl    *UploadURL
	}{
		{int64(1), "1", "", ts, &UploadURL{Success: Bool(true), URL: String("https://s3-transfer-1")}},
		{int64(2), "2", "", ts, &UploadURL{Success: Bool(true), URL: String("https://s3-transfer-2")}},
		{int64(1), "1", "xx", board, &UploadURL{Success: Bool(true), URL: String("https://s3-board-1")}},
		{int64(2), "2", "xx", board, &UploadURL{Success: Bool(true), URL: String("https://s3-board-2")}},
	}

	for _, tt := range tests {
		var path string
		switch tt.bot.(type) {
		case *Transfer:
			path = fmt.Sprintf("/transfers/%v/files/%v/upload-url/%v", tt.bot.GetID(), tt.fid, tt.partNum)
		case *Board:
			path = fmt.Sprintf("/boards/%v/files/%v/upload-url/%v/%v", tt.bot.GetID(), tt.fid, tt.partNum, tt.mid)
		}
		func(urlStr string) {
			mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, `{"success": true, "url": "%v"}`, urlStr)
			})
		}(*tt.uurl.URL)
	}

	for _, tt := range tests {
		got, _ := client.uploader.getUploadURL(context.Background(), tt.bot, tt.fid, tt.partNum, tt.mid)
		if !reflect.DeepEqual(got, tt.uurl) {
			t.Errorf("Transfers.GetUploadURL returned %+v, want %+v", got, tt.uurl)
		}
	}
}

func TestUploadBytes(t *testing.T) {
	_, s3, s3url, teardown := setup()
	defer teardown()

	s3path := "/p/1"

	s3.HandleFunc(s3path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	uurl := &UploadURL{
		Success: Bool(true),
		URL:     String(s3url + s3path),
	}

	err := uploadBytes(context.Background(), uurl, []byte("pony data"))
	if err != nil {
		t.Errorf("uploadBytes returned an error: %v", err)
	}
}

func TestUploadBytes_noSuchKey(t *testing.T) {
	_, s3, s3url, teardown := setup()
	defer teardown()

	s3.HandleFunc("/not/found/file/1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprint(w, `
			<?xml version="1.0" encoding="UTF-8"?>
			<Error>
			  <Code>NoSuchKey</Code>
			  <Message>The resource you requested does not exist</Message>
			  <Resource>/mybucket/myfoto.jpg</Resource> 
			  <RequestId>4442587FB7D0A2F9</RequestId>
			</Error>
		`)
	})

	uurl := &UploadURL{
		Success: Bool(true),
		URL:     String(s3url + "/not/found/file/1"),
	}

	err := uploadBytes(context.Background(), uurl, []byte("pony data"))

	if err == nil {
		t.Errorf("Expected error to be returned")
	}
}

func TestUploaderService_getUploadURL_fail(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	tests := []struct {
		url       string
		httpCode  int
		partNum   int64
		wantError string
	}{
		{"/transfers/2/files/2/upload-url/1", 404, int64(1), "Invalid transfer or file id."},
		{"/transfers/2/files/2/upload-url/0", 417, int64(0), "Chunk numbers are 1-based"},
	}

	ts := &Transfer{ID: String("2")}

	for _, g := range tests {
		mux.HandleFunc(g.url, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(g.httpCode)
			fmt.Fprintf(w, `{"success":false,"message":"%v"}`, g.wantError)
		})
		_, err := client.uploader.getUploadURL(context.Background(), ts, "2", g.partNum, "")
		testErrorResponse(t, err, g.wantError)
	}
}
