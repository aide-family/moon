package bo

import "time"

// TaskSnapshot 是领取到的待投递任务。
type TaskSnapshot struct {
	UID             string
	NamespaceUID    string
	Phase           string
	ChannelUID      string
	ChannelType     string
	ChannelApp      string
	ChannelEndpoint string
	ChannelUsername string
	ChannelPassword string
	ChannelSecret   string
	ChannelFrom     string
	Sender          string
	TemplateUID     string
	TemplateType    string
	TemplateSubject string
	TemplateBody    string
	Receivers       string
	Params          string
	Subject         string
	Body            string
}

// WorkerStatus is the safe instance heartbeat sent to Alioth.
type WorkerStatus struct {
	InstanceUID   string
	Hostname      string
	Busy          bool
	TaskUID       string
	TaskName      string
	TaskStartedAt time.Time
}

// NotifyReport 是投递回写。
type NotifyReport struct {
	TaskUID string
	OK      bool
	Detail  string
	Subject string
	Content string
}
