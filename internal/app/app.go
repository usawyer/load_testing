package app

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/usawyer/load_testing/configs"
	"github.com/usawyer/load_testing/internal/test_client"
	"golang.org/x/time/rate"
)

func Run(cfg *configs.Config) {
	limiter := rate.NewLimiter(rate.Limit(cfg.HTTP.MaxRPC), 1)
	counter := 0

	var wg sync.WaitGroup
	wg.Add(cfg.HTTP.ThreadsNum)

	for i := 0; i < cfg.HTTP.ThreadsNum; i++ {
		go func() {
			defer wg.Done()

			client, err := test_client.New(limiter, i+1)
			if err != nil {
				slog.Error(fmt.Sprintf("Test failed: %v", err))
				return
			}

			if err := client.InitTest(cfg.HTTP.StartPage); err != nil {
				slog.Error(fmt.Sprintf("Test failed: %v", err))
			} else {
				counter++
				slog.Info(fmt.Sprintf("Test for client №%d successfully passed", i+1))
			}
		}()
	}

	wg.Wait()
	slog.Info(fmt.Sprintf("%d of %d tests was successfully passed", counter, cfg.HTTP.ThreadsNum))
}
