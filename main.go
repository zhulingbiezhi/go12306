package main

import (
	"context"

	"github.com/zhulingbiezhi/go12306/pkg/job"
	"github.com/zhulingbiezhi/go12306/tools/logger"
)

func main() {
	for _, job := range job.Jobs {
		err := job.Run(context.Background())
		if err != nil {
			logger.Error(err)
		}
	}
	select {}
}
