package grpcapi

import (
	"context"
	"errors"
	"movie_downloader/homework-3/internal/filmclient"
	moviepb "movie_downloader/homework-3/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
    moviepb.UnimplementedMovieServiceServer
    client *filmclient.Client
}

func NewServer(client *filmclient.Client) *Server  {
	server := Server {
		client: client,
	}
	return &server
} 


func (s *Server) GetLastMovie( 
	ctx context.Context,
	req *moviepb.GetLastMovieRequest,
)(*moviepb.Movie, error) {
	film, err := s.client.GetLastMovie(ctx)

	if errors.Is(err, filmclient.ErrMovieNotFound) {
		return nil, status.Error(codes.NotFound, "movie not found")
	}

	if errors.Is(err, context.Canceled) {
    	return nil, status.Error(codes.Canceled, "request canceled")
	}
	
	if errors.Is(err, context.DeadlineExceeded) {
    	return nil, status.Error(codes.DeadlineExceeded, "request deadline exceeded")
	}

	if err != nil {
    	return nil, status.Error(codes.Internal, "failed to get last movie")
	}

	response := &moviepb.Movie{
    	Id:    int32(film.ID),
    	Title: film.Title,
		Year: int32(film.Year),
		Director: film.Director,
	}
	return response, nil
}

func (s *Server) LoadMovies( 
	ctx context.Context,
	req *moviepb.LoadMoviesRequest,
)(*moviepb.LoadMoviesResponse, error) {

	var pbFilms []*moviepb.Movie

	if req.GetFrom() < 1 || req.GetTo() < req.GetFrom() {
    	return nil, status.Error(codes.InvalidArgument,"invalid movie range")
	}

	films, err := s.client.LoadMovies(ctx, int(req.From), int(req.To))

	if errors.Is(err, context.Canceled) {
    	return nil, status.Error(codes.Canceled, "request canceled")
	}

	if errors.Is(err, context.DeadlineExceeded) {
    	return nil, status.Error(codes.DeadlineExceeded, "request deadline exceeded")
	}

	if err != nil && len(films) == 0 {
		return nil, status.Error(codes.Unavailable,"failed to load movies")
	}

	for _, film := range films {
		pbFilm := &moviepb.Movie{
			Id: int32(film.ID),
			Year: int32(film.Year),
			Title: film.Title,
			Director: film.Director,
		}
		pbFilms = append(pbFilms, pbFilm)
	}

	response := &moviepb.LoadMoviesResponse{
		Movies: pbFilms,
	}

	if err != nil {
		response.Error = err.Error()
	}
	return response, nil
}