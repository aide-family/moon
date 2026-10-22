package auth

import (
	"context"
	"log/slog"
	"testing"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	megrezv1 "github.com/aide-family/Alioth/api/megrez/v1"
	merakv1 "github.com/aide-family/Alioth/api/merak/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/actor"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

type userStub struct{ repository.UserRepo }

func (userStub) UserByUID(context.Context, string) (*bo.Account, error) {
	return &bo.Account{UID: "user-1", Status: "ACTIVE"}, nil
}

func TestClaimRejectsUserJWT(t *testing.T) {
	authenticator, err := NewAuthenticator(userStub{}, nil, nil, nil, nil, nil, &conf.Auth{
		JwtSecret:   "test-secret",
		ServiceKeys: []*conf.Auth_ServiceKey{{Name: "megrez", Key: "sk-megrez-dev"}},
	}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	token, err := authenticator.sign("user-1", "access", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	_, err = authenticator.Authenticate(context.Background(), megrezv1.EvalService_ClaimRules_FullMethodName, "Bearer "+token, "")
	if kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("user jwt claim: %v", err)
	}
	ctx, err := authenticator.Authenticate(context.Background(), megrezv1.EvalService_ReportEval_FullMethodName, "Bearer sk-megrez-dev", "")
	if err != nil {
		t.Fatal(err)
	}
	name, ok := actor.ServiceName(ctx)
	if !ok || name != "megrez" {
		t.Fatalf("service = %q ok=%v", name, ok)
	}
	statusCtx, err := authenticator.Authenticate(context.Background(), megrezv1.EvalService_ReportWorkerStatus_FullMethodName, "Bearer sk-megrez-dev", "")
	if err != nil {
		t.Fatalf("megrez status report: %v", err)
	}
	if name, ok := actor.ServiceName(statusCtx); !ok || name != "megrez" {
		t.Fatalf("status reporter = %q ok=%v", name, ok)
	}
	_, err = authenticator.Authenticate(context.Background(), merakv1.NotifyService_ReportWorkerStatus_FullMethodName, "Bearer sk-megrez-dev", "")
	if kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("megrez key reached merak heartbeat: %v", err)
	}
	_, err = authenticator.Authenticate(context.Background(), v1.OperationAlertEventServiceListAlertEvents, "Bearer sk-megrez-dev", "ns")
	if kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("service key list: %v", err)
	}
}
