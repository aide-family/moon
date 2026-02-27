// Package job is the job command for the Rabbit service
package job

import (
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"

	"github.com/aide-family/rabbit/cmd"
	"github.com/aide-family/rabbit/cmd/run"
)

const cmdJobLong = `Start the Rabbit job service (EventBus) only, providing asynchronous message processing capabilities.

The job command starts only the EventBus/Job service component, which handles:
  • Message queue processing: Process messages from the queue asynchronously
  • Background tasks: Handle background message delivery tasks (email, Webhook, SMS, etc.)
  • Event-driven processing: Process messages based on events and triggers
  • Worker pool management: Manage worker pools for concurrent message processing

Key Features:
  • Asynchronous processing: Process messages asynchronously from the queue without blocking API calls
  • Configurable worker pool: Adjustable worker pool size via --job-core-worker-total for optimal performance
  • Event-driven architecture: Event-driven message processing for scalable and resilient message delivery
  • High throughput: Support for high-throughput message processing with configurable buffer size
  • Timeout control: Configurable timeout for message processing via --job-core-timeout

Use Cases:
  • Message queue processing: Deploy job service separately for dedicated message queue processing
  • Background tasks: Handle background message delivery tasks independently from API services
  • Scalability: Scale job service independently for better performance and resource utilization
  • Microservices: Deploy job service as a separate microservice in distributed architectures
  • High-volume processing: Dedicate resources to message processing for high-volume scenarios

Note: This command only starts the job service. For API access (HTTP/gRPC), you need to start the
http or grpc service separately. The job service processes messages that are submitted through
the HTTP or gRPC APIs.

After starting the service, Rabbit job will:
  • Listen on the configured job port (default: 0.0.0.0:17070, configurable via --job-address)
  • Start processing messages from the queue asynchronously
  • Handle background message delivery tasks with the configured worker pool`

func NewCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "job",
		Short: "Run the Rabbit job service (EventBus) only",
		Long:  cmdJobLong,
		Annotations: map[string]string{
			"group": cmd.ServiceCommands,
		},
		Run: func(_ *cobra.Command, _ []string) {
			if err := flags.applyToBootstrap(); err != nil {
				klog.Errorw("msg", "apply to bootstrap failed", "error", err)
				return
			}
			run.NewEngine(run.NewEndpoint(WireApp)).Start()
		},
	}

	flags.addFlags(runCmd)
	return runCmd
}
