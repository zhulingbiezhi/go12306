package main

import (
	"github.com/zhulingbiezhi/go12306/config"
	"github.com/zhulingbiezhi/go12306/pkg/account"
	"github.com/zhulingbiezhi/go12306/pkg/job"
)

func main() {
	cfg, err := config.LoadConf()
	if err != nil {
		panic(err)
	}
	jobMgr := &job.JobMgr{}
	userMgr := &account.AccountMgr{}
	err = userMgr.Init(cfg.Accounts)
	if err != nil {
		panic(err)
	}
	err = jobMgr.Init(userMgr, cfg.Jobs)
	if err != nil {
		panic(err)
	}
	err = jobMgr.Run()
	if err != nil {
		panic(err)
	}
	select {}
}
