package filmclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"movie_downloader/homework-3/internal/movie"
	"net/http"
	"time"
)

var ErrMovieNotFound = errors.New("movie not found")

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	client := Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	return &client
}

func (c *Client) GetMovie(ctx context.Context, id int) (*movie.Movie, error) {
	url := fmt.Sprintf("%s/%d/info.0.json", c.baseURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrMovieNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var film movie.Movie

	err = json.NewDecoder(resp.Body).Decode(&film)

	if err != nil {
		return nil, err
	}

	return &film, nil

}

func (c *Client) LoadMovies(ctx context.Context, from int, to int) ([]movie.Movie, error) {

	if from < 1 || to < from {
		return nil, fmt.Errorf("invalid movie range: from=%d, to=%d", from, to)
	}

	var movies []movie.Movie
	var errs []error

	for id := from; id <= to; id++ {
		film, err := c.GetMovie(ctx, id)

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return movies, errors.Join(append(errs, err)...)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("movie %d: %w", id, err))
			continue
		}
		movies = append(movies, *film)
	}

	return movies, errors.Join(errs...)
}

func (c *Client) GetLastMovie(ctx context.Context) (*movie.Movie, error) {
	var lastFilm *movie.Movie
	maxInt := int(^uint(0) >> 1)
	left := 0
	right := 1

	for {
		film, err := c.GetMovie(ctx, right)

		if errors.Is(err, ErrMovieNotFound) {
			break
		}

		if err != nil {
			return nil, err
		}

		if right == maxInt {
			return film, nil
		}

		left = right

		if right > maxInt/2 {
			right = maxInt
		} else {
			right *= 2
		}
		lastFilm = film

	}

	for (right - left) > 1 {

		mid := left + (right-left)/2
		film, err := c.GetMovie(ctx, mid)

		if errors.Is(err, ErrMovieNotFound) {
			right = mid

		} else if err != nil {
			return nil, err
		} else {
			left = mid
			lastFilm = film
		}
	}

	if left == 0 {
		return nil, ErrMovieNotFound
	}

	return lastFilm, nil
}


