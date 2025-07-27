export enum WorkflowStatus {
	/**
	 *   StatusPending indicates the workflow is pending and has not started.
	*/
	StatusPending = "pending",
	/**
	 *   StatusRunning indicates the workflow is currently running.
	*/
	StatusRunning = "running",
	/**
	 *   StatusCompleted indicates the workflow has completed successfully.
	*/
	StatusCompleted = "completed",
	/**
	 *   StatusFailed indicates the workflow has failed.
	*/
	StatusFailed = "failed"
}