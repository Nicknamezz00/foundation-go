## Usage

```go
cron.InitStart([]*cron.Job{
  {
    Spec: "@every 60s",
    Command: cron.JobWrapper("fetch", func(ctx context.Context) error {)
      // do something
      return nil
    }),
  }
})
```