package main

import (
	"context"
	"go12306/helpers/logger"
	"go12306/job"
	_ "go12306/job"
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
