package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Movie struct { // структура одного фильма
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Director string `json:"director"`
}

func worker(ctx context.Context, jobs chan int, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()

	for {
		select {

		case <-ctx.Done():
			fmt.Println("Worker получил отмену")
			return

		case id, ok := <-jobs:

			if !ok {
				return
			}

			movie, err := downloadMovie(ctx, id, timeout)
			if err != nil {
				fmt.Printf("Ошибка обработки фильма %d: %v\n", id, err)
				continue
			}
			fmt.Printf("%d - %s - %d - %s\n", movie.ID, movie.Title, movie.Year, movie.Director)
		}

	}
}

func downloadMovie(ctx context.Context, id int, timeout time.Duration) (Movie, error) {

	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	url := fmt.Sprintf("https://homeworksite.site/%d/info.0.json", id)

	req, err := http.NewRequestWithContext(requestCtx, "GET", url, nil)

	if err != nil {
		return Movie{}, fmt.Errorf("ошибка создания HTTP-запроса: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return Movie{}, fmt.Errorf("ошибка HTTP запроса: %w", err)
	}

	defer resp.Body.Close()

	cd := resp.StatusCode

	if cd != 200 {
		return Movie{}, fmt.Errorf("сервер вернул ошибку со статусом: %d", cd)
	}

	var movie Movie

	jd := json.NewDecoder(resp.Body)

	err = jd.Decode(&movie)

	if err != nil {
		return Movie{}, err
	}

	return movie, nil
}

func main() {

	fromSet := false
	toSet := false

	from := flag.Int("from", 0, "начальный ID фильма")
	to := flag.Int("to", 0, "конечный ID фильма")
	workers := flag.Int("workers", 10, "кол-во воркеров")
	timeout := flag.Duration("timeout", 5*time.Second, "таймаут HTTP запроса")
	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "from" {
			fromSet = true
		}
		if f.Name == "to" {
			toSet = true
		}
	})

	if fromSet == false || toSet == false {
		fmt.Println("отсутствует обязательный параметр")
		os.Exit(1)
	}

	if *from > *to {
		fmt.Println("ошибка диапазона")
		os.Exit(1)
	}

	if *workers <= 0 {
		fmt.Println("некорректное кол-во воркеров")
		os.Exit(1)
	}

	var wg sync.WaitGroup

	jobs := make(chan int)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	wg.Add(*workers)

	for i := 1; i <= *workers; i++ {
		go worker(ctx, jobs, &wg, *timeout)
	}

sendJobs:
	for i := *from; i <= *to; i++ {
		select {
		case <-ctx.Done():
			break sendJobs
		case jobs <- i:

		}
	}

	close(jobs)

	wg.Wait()

}
