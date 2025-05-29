// Package cron_server provides a flexible and robust cron job scheduling system.
// It implements the transport.Server interface from go-kratos framework and offers
// a convenient way to manage and execute scheduled tasks.
//
// The package includes several predefined cron specifications:
//   - @yearly/@annually: Run once a year
//   - @monthly: Run once a month
//   - @weekly: Run once a week
//   - @daily/@midnight: Run once a day
//   - @hourly: Run once an hour
//
// You can also create custom cron specifications using CronSpecCustom or
// interval-based specifications using CronSpecEvery.
//
// Example usage:
//
//	server := cron_server.NewCronJobServer("my-cron", logger)
//	job := &MyJob{
//		spec: cron_server.CronSpecDaily,
//		index: "daily-task",
//	}
//	server.AddJob(job)
//	server.Start(context.Background())
//
// The CronJobServer provides methods to:
//   - Add new jobs (AddJob)
//   - Force update existing jobs (AddJobForce)
//   - Remove jobs (RemoveJob)
//   - Start and stop the server (Start, Stop)
//
// Each job must implement the CronJob interface, which requires:
//   - ID(): Return the job's entry ID
//   - Index(): Return a unique identifier for the job
//   - Spec(): Return the cron specification
//   - WithID(): Set the job's entry ID
//   - IsImmediate(): Determine if the job should run immediately
//   - Run(): Implement the actual job logic
package cron_server
