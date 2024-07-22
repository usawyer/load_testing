package app

import (
	"fmt"
	"github.com/usawyer/load_testing/configs"
	"github.com/usawyer/load_testing/internal/test"
	"log/slog"
	"sync"
)

func Run(cfg *configs.Config) {
	var wg sync.WaitGroup
	wg.Add(cfg.HTTP.ThreadsNum)

	for i := 0; i < cfg.HTTP.ThreadsNum; i++ {
		go func() {
			defer wg.Done()

			client, err := test.New(cfg.HTTP.MaxRPC, i+1)
			if err != nil {
				slog.Error(fmt.Sprintf("Test failed: %v", err))
				return
			}

			if err := client.InitTest(cfg.HTTP.StartPage); err != nil {
				slog.Error(fmt.Sprintf("Test failed: %v", err))
			} else {
				slog.Info(fmt.Sprintf("Test for client №%d successfully passed", i+1))
			}
		}()
	}

	wg.Wait()
}
