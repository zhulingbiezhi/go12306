package main

import (
	"context"
	"github.com/zhulingbiezhi/go12306/helpers/logger"
	"github.com/zhulingbiezhi/go12306/job"
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
