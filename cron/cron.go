package cron

import (
	"context"
	"foundation-go/log"

	"github.com/robfig/cron/v3"
)

type Handler struct {
	Cron *cron.Cron
}

type Job struct {
	Spec    string
	Command func()
}

func (h *Handler) Start() {
	h.Cron.Start()
}

func (h *Handler) Stop() {
	h.Cron.Stop()
}

func (h *Handler) Register(jobs []*Job) {
	for _, job := range jobs {
		if _, err := h.Cron.AddFunc(job.Spec, job.Command); err != nil {
			log.Fatalf(context.Background(), log.Fields{log.Error: err}, "%s", "_register_cron_job_failed")
		}
	}
}

func InitStart(jobs []*Job) {
	h := &Handler{
		Cron: cron.New(cron.WithSeconds()),
	}
	h.Register(jobs)
	h.Start()
}

type cmd func(ctx context.Context) error

func JobWrapper(key string, f cmd) func() {
	return func() {
		ctx := context.Background()
		err := f(ctx)
		logM := log.Fields{"key": key}
		if err != nil {
			logM[log.Error] = err
			log.Errorf(ctx, logM, "%s", "_cron_job_failed")
			return
		}
		log.Infof(ctx, logM, "%s", "_cron_job_succeed")
	}
}
