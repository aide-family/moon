package controller

import "time"

type Interface interface {

	// Initial indicates that the cluster is in the initialization state.
	// During this phase, the cluster's connection configuration will be checked for availability.
	// If the cluster initialization fails, it will remain in this phase.
	// otherwise, the status will be advanced to the next phase: Running.
	Initial(c *Context) (*time.Duration, error)

	// Running indicates that the cluster is in a running state.
	// During this phase, cluster information will be synchronized to the status, such as the API list, node information, etc.
	// When the cluster is disabled, the condition ClusterCondDisabled status will be updated to True, indicating the cluster is disabled.
	// When the cluster condition ClusterCondDisabled status is False, and the connection to the cluster is successful,
	// the cluster condition ClusterCondReady status will be updated to True; otherwise, it will be updated to False.
	// When updating the cluster condition ClusterCondReady to False, depending on the cause, the Reason will be set to Disconnected or Disabled.
	Running(c *Context) (*time.Duration, error)

	// Terminating will release this cluster.
	// When the cluster is being deleted, the status will be advanced to Terminating.
	// At termination, some finalizing actions will be taken for the cluster, and then the Finalizer will be removed.
	Terminating(c *Context) (*time.Duration, error)
}
