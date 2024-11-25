package main

import (
	"log"
	"net"
	"net/http"
	"path/filepath"
	"os"
	"fmt"
	"io"

	"github.com/kmin1231/proj_grpc/pkg/grpcserver"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    file, header, err := r.FormFile("video")
    if err != nil {
        http.Error(w, "Error reading file", http.StatusInternalServerError)
        return
    }
    defer file.Close()

    filePath := filepath.Join("videos", header.Filename)
    outFile, err := os.Create(filePath)
    if err != nil {
        http.Error(w, "Error saving file", http.StatusInternalServerError)
        return
    }
    defer outFile.Close()

    _, err = io.Copy(outFile, file)
    if err != nil {
        http.Error(w, "Error saving file", http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "File uploaded successfully")
}


func main() {
	videoDir, _ := filepath.Abs("./videos")

	grpcServer := grpcserver.NewServer(videoDir)
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		log.Println("Starting gRPC server on port 50051...")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Ensure the videos directory exists
    if _, err := os.Stat("videos"); os.IsNotExist(err) {
        os.Mkdir("videos", os.ModePerm)
    }


	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/videos", grpcserver.HandleVideoList(videoDir))
	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		videoName := r.URL.Query().Get("video")
		log.Printf("Streaming video: %s\n", videoName)
		grpcserver.HandleVideoStream(videoDir)(w, r)
	})
	
	http.Handle("/", http.FileServer(http.Dir("./web")))

	log.Println("Starting HTTP server on port 9000...")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatalf("Failed to serve HTTP: %v", err)
	}
}
