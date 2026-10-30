package main

import (
	"movie_downloader/homework-3/internal/filmclient"
	"movie_downloader/homework-3/internal/httpapi"
	"net/http"
	"log"
	"os"
	"net"
	"google.golang.org/grpc"
	"movie_downloader/homework-3/internal/grpcapi"
	moviepb "movie_downloader/homework-3/proto"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "movie_downloader/homework-3/docs"
)


// @title Movie API
// @version 1.0
// @description HTTP API для загрузки фильмов и поиска последнего фильма.
// @host localhost:8080
// @BasePath /
func main() {
	baseURL := os.Getenv("MOVIE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://homeworksite.site"
	}

	client := filmclient.NewClient(baseURL)
	handler := httpapi.NewHandler(client)
	mux := handler.Router()
	mux.Handle("GET /swagger/", httpSwagger.Handler())
	
	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	grpcAddr := os.Getenv("GRPC_ADDR")

	if grpcAddr == "" {
		grpcAddr = ":50051"
	}

	listener, err := net.Listen("tcp", grpcAddr)

	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	movieServer := grpcapi.NewServer(client)
	moviepb.RegisterMovieServiceServer(grpcServer, movieServer)

	go func() {
    err := grpcServer.Serve(listener)
    if err != nil {
        log.Fatal(err)
    }
	}()

	err = http.ListenAndServe(httpAddr, mux)
	if err != nil {
    	log.Fatal(err)
	}	
}	
