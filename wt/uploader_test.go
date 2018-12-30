package wt

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestUploaderService_send(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	s3, s3url, s3teardown := setupS3()
	defer s3teardown()

	tests := []struct {
		partNumbers int
		chunkSize   int
		fileID      string
		idx         identifiable
	}{
		{2, 1024, "1", &Transfer{ID: String("1")}},
		{2, 256, "2", &Board{ID: String("2")}},
	}

	for _, test := range tests {
		var prefix string
		switch test.idx.(type) {
		case *Transfer:
			prefix = "transfers"
		case *Board:
			prefix = "boards"
		}
		for i := 1; i <= test.partNumbers; i++ {
			func(partNum int) {
				path := fmt.Sprintf("/%v/%v/files/%v/upload-url/%v", prefix, test.idx.GetID(), test.fileID, partNum)
				mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintf(w, `{"success": true, "url": "%s"}`, fmt.Sprintf("%v/%v/p/%v", s3url, prefix, partNum))
				})
				s3.HandleFunc(fmt.Sprintf("/%v/p/%v", prefix, partNum), func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
				})
			}(i)
		}

		file := &File{
			ID:   &test.fileID,
			Name: String("pony.txt"),
			Multipart: &Multipart{
				PartNumbers: Int64(int64(test.partNumbers)),
				ChunkSize:   Int64(int64(test.chunkSize)),
			},
		}

		data := make([]byte, test.partNumbers*test.chunkSize)
		for i := range data {
			data[i] = 'x'
		}
		buf := NewBuffer("pony.txt", data)
		ft := newFileTransfer(buf, file)
		err := client.uploader.send(context.Background(), test.idx, ft)
		if err != nil {
			t.Errorf("send returned an error: %+v", err)
		}
	}
}

func TestUploaderService_getUploadURL(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/transfers/1/files/1/upload-url/1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"success": true, "url": "https://s3-1"}`)
	})
	mux.HandleFunc("/transfers/1/files/1/upload-url/2", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"success": true, "url": "https://s3-2"}`)
	})

	tests := []struct {
		in  int64
		out *UploadURL
	}{
		{int64(1), &UploadURL{Success: Bool(true), URL: String("https://s3-1")}},
		{int64(2), &UploadURL{Success: Bool(true), URL: String("https://s3-2")}},
	}

	ts := &Transfer{ID: String("1")}

	for _, test := range tests {
		got, _ := client.uploader.getUploadURL(context.Background(), ts, "1", test.in)
		if !reflect.DeepEqual(got, test.out) {
			t.Errorf("Transfers.GetUploadURL returned %+v, want %+v", got, test.out)
		}
	}
}

func TestUploadBytes(t *testing.T) {
	s3, s3url, teardown := setupS3()
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
	s3, s3url, teardown := setupS3()
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
		_, err := client.uploader.getUploadURL(context.Background(), ts, "2", g.partNum)
		testErrorResponse(t, err, g.wantError)
	}
}
