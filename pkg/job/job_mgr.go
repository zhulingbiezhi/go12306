package job

import (
	"context"

	"github.com/zhulingbiezhi/go12306/config"
	"github.com/zhulingbiezhi/go12306/pkg/account"
	"github.com/zhulingbiezhi/go12306/tools/errors"
)

type JobMgr struct {
	jobs map[string]*UserJob
}

func (mgr *JobMgr) GetJob(name string) (*UserJob, error) {
	job, ok := mgr.jobs[name]
	if !ok {
		return nil, errors.Errorf(nil, "job name: %s is empty", name)
	}
	return job, nil
}

func (mgr *JobMgr) Init(accMgr *account.AccountMgr, cfgs []*config.Job) error {
	mgr.jobs = make(map[string]*UserJob)
	for _, cfg := range cfgs {
		acc, err := accMgr.GetAccount(cfg.Account)
		if err != nil {
			return err
		}
		job := &UserJob{
			Name:               cfg.Name,
			Account:            acc,
			Seats:              cfg.Seats,
			ExceptTrainNumbers: cfg.ExceptTrainNumbers,
			TrainNumbers:       cfg.TrainNumbers,
			Dates:              cfg.Dates,
			Station:            cfg.Station,
			Members:            cfg.Members,
			QueryInterval:      cfg.QueryInterval,
			AllowLessMember:    cfg.AllowLessMember,
		}
		_, ok := mgr.jobs[cfg.Name]
		if ok {
			return errors.Errorf(nil, "job name: %s is exist", cfg.Name)
		}
		mgr.jobs[cfg.Name] = job
	}
	return nil
}

func (mgr *JobMgr) Run() error {
	for _, job := range mgr.jobs {
		err := job.Run(context.Background())
		if err != nil {
			return err
		}
	}
	return nil
}
