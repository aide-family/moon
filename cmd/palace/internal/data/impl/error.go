package impl

import (
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"

	"github.com/aide-family/moon/pkg/merr"
)

func userNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorUserNotFound("user not found").WithCause(err)
	}
	return err
}

func oauthUserNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorUserNotFound("oauth user not found").WithCause(err)
	}
	return err
}

func teamDashboardNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("team dashboard not found").WithCause(err)
	}
	return err
}

func teamDashboardChartNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("team dashboard chart not found").WithCause(err)
	}
	return err
}

func teamNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("team not found").WithCause(err)
	}
	return err
}

func teamMemberNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("team member not found").WithCause(err)
	}
	return err
}

func menuNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("menu not found").WithCause(err)
	}
	return err
}

func strategyNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("strategy not found").WithCause(err)
	}
	return err
}

func strategyMetricNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("metric strategy not found").WithCause(err)
	}
	return err
}

func strategyGroupNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("strategy group not found").WithCause(err)
	}
	return err
}

func teamRoleNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("team role not found").WithCause(err)
	}
	return err
}

func noticeGroupNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("notice group not found").WithCause(err)
	}
	return err
}

func hookNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("hook not found").WithCause(err)
	}
	return err
}

func teamDictNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("team dict not found").WithCause(err)
	}
	return err
}

func datasourceNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("datasource not found").WithCause(err)
	}
	return err
}

func teamSMSConfigNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("team sms config not found").WithCause(err)
	}
	return err
}

func teamEmailConfigNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("team email config not found").WithCause(err)
	}
	return err
}

func roleNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("role not found").WithCause(err)
	}
	return err
}

func auditNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("audit not found").WithCause(err)
	}
	return err
}

func sendMessageLogNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("send message log not found").WithCause(err)
	}
	return err
}

func realtimeNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("realtime not found").WithCause(err)
	}
	return err
}

func teamTimeEngineNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("team time engine not found").WithCause(err)
	}
	return err
}

func teamTimeEngineRuleNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("team time engine rule not found").WithCause(err)
	}
	return err
}

func teamDatasourceMetricMetadataNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("team datasource metric metadata not found").WithCause(err)
	}
	return err
}

func strategyMetricRuleNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return merr.ErrorNotFound("strategy metric rule not found").WithCause(err)
	}
	return err
}
