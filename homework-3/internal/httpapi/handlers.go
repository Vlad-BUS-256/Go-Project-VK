package httpapi

import (
	"encoding/json"
	"errors"
	"movie_downloader/homework-3/internal/filmclient"
	"movie_downloader/homework-3/internal/movie"
	"net/http"
	"strconv"
	"context"
)

type PartialMoviesResponse struct {
	Movies []movie.Movie `json:"movies"`
	Error  string        `json:"error"`
}

type Handler struct {
	client *filmclient.Client
}

func NewHandler(client *filmclient.Client) *Handler {
	handler := Handler{
		client: client,
	}

	return &handler
}


// GetLastMovie godoc
// @Summary Получить последний фильм
// @Description Находит фильм с максимальным существующим ID
// @Tags movies
// @Produce json
// @Success 200 {object} movie.Movie
// @Failure 404 {string} string "Movie not found"
// @Failure 502 {string} string "Upstream error"
// @Router /last-movie [get]
func (h *Handler) GetLastMovie(w http.ResponseWriter, r *http.Request) {
	film, err := h.client.GetLastMovie(r.Context())

	if errors.Is(err, context.Canceled) {
    	return
	}
	
	if errors.Is(err, filmclient.ErrMovieNotFound) {
		http.Error(w, "movie not found", http.StatusNotFound)
		return
	}

	if errors.Is(err, context.DeadlineExceeded) {
    	http.Error(w, "request deadline exceeded", http.StatusGatewayTimeout)
    	return
	}

	if err != nil {
		http.Error(w, "failed to fetch movie", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(film)

}

// LoadMovies godoc
// @Summary Загрузить фильмы по диапазону ID
// @Description Загружает фильмы от from до to включительно
// @Tags movies
// @Produce json
// @Param from query int true "Начальный ID фильма"
// @Param to query int true "Конечный ID фильма"
// @Success 200 {array} movie.Movie
// @Failure 400 {string} string "Invalid movie range"
// @Failure 502 {object} PartialMoviesResponse "Failed to load movies"
// @Router /load-movies [get]
func (h *Handler) LoadMovies(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, err := strconv.Atoi(fromStr)

	if err != nil {
		http.Error(w, "invalid from parameter", http.StatusBadRequest)
		return
	}

	to, err := strconv.Atoi(toStr)
	if err != nil {
		http.Error(w, "invalid to parameter", http.StatusBadRequest)
		return
	}

	if from < 1 || to < from {
		http.Error(w, "invalid movie range", http.StatusBadRequest)
		return
	}

	films, err := h.client.LoadMovies(r.Context(), from, to)

	if errors.Is(err, context.DeadlineExceeded) {
    	http.Error(w, "request deadline exceeded", http.StatusGatewayTimeout)
    	return
	}

	if errors.Is(err, context.Canceled) {
    	return
	}

	if err != nil && len(films) > 0 {
		response := PartialMoviesResponse{
			Movies: films,
			Error:  err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(response)
		return
	}

	if err != nil {
		http.Error(w, "failed to load movies", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(films)
}

func (h *Handler) Router() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /last-movie", h.GetLastMovie)
	mux.HandleFunc("GET /load-movies", h.LoadMovies)
	return mux
}
