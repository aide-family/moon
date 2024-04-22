package terminal

import (
	"context"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"net/http"
)

type PodOptions struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	Container string `json:"container,omitempty"`
	Shell     string `json:"shell,omitempty"`
}

type VMOptions struct {
	Namespace   string `json:"namespace,omitempty"`
	Name        string `json:"name,omitempty"`
	Subresource string `json:"subresource,omitempty"`
}

func PodWsHandler(ctx context.Context, cli *rest.RESTClient, config *rest.Config, opt PodOptions, w http.ResponseWriter, r *http.Request) error {
	session, err := NewTerminalSession(w, r)
	if err != nil {
		return err
	}
	defer func() {
		_ = session.Close()
	}()
	req := cli.Post().
		Resource("pods").
		Namespace(opt.Namespace).
		Name(opt.Name).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: opt.Container,
			Command:   []string{opt.Shell},
			Stderr:    true,
			Stdin:     true,
			Stdout:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, http.MethodPost, req.URL())
	if err != nil {
		return err
	}
	if err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             session,
		Stdout:            session,
		Stderr:            session,
		Tty:               true,
		TerminalSizeQueue: session,
	}); err != nil {
		_, _ = session.Write([]byte("exec pod command failed," + err.Error()))
		session.Done()
	}
	return nil
}
