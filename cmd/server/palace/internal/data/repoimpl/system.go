package repoimpl

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
)

// NewSystemRepository 创建系统相关功能实现
func NewSystemRepository(data *data.Data) repository.System {
	return &systemRepositoryImpl{data: data}
}

type systemRepositoryImpl struct {
	data *data.Data
}

// DeleteBackup 删除备份数据库
func (s *systemRepositoryImpl) DeleteBackup(ctx context.Context, teamID uint32) {
	databaseName := s.getBackupTeamDatabaseName(teamID)
	_ = s.resetTeam(ctx, databaseName)
}

// RestoreData 恢复团队数据
func (s *systemRepositoryImpl) RestoreData(ctx context.Context, teamID uint32) error {
	return s.restoreData(ctx, teamID)
}

// ResetTeam 重置团队数据库
func (s *systemRepositoryImpl) ResetTeam(ctx context.Context, teamID uint32) error {
	databaseName := s.getBackupTeamDatabaseName(teamID)
	oldDatabaseName := data.GenBizDatabaseName(teamID)
	// 备份数据库数据
	if err := s.backupTeam(ctx, databaseName, oldDatabaseName); err != nil {
		return err
	}

	// 删除数据库
	if err := s.resetTeam(ctx, oldDatabaseName); err != nil {
		return err
	}
	return nil
}

// getBackupTeamDatabaseName 获取备份团队数据库名称
func (s *systemRepositoryImpl) getBackupTeamDatabaseName(teamID uint32) string {
	return "biz_backup_team_" + strconv.FormatUint(uint64(teamID), 10)
}

// backupTeam 备份团队数据
func (s *systemRepositoryImpl) backupTeam(ctx context.Context, databaseName, oldDatabaseName string) (err error) {
	defer func() {
		if !types.IsNil(err) {
			err = merr.ErrorAlert("备份团队数据失败").WithCause(err)
			// 删除备份数据库
			_, err = s.data.GetBizDB(ctx).Exec("DROP DATABASE IF EXISTS `" + databaseName + "`")
			if !types.IsNil(err) {
				log.Errorf("删除备份数据库失败: %v", err)
			}
		}
	}()
	// 创建备份数据库
	_, err = s.data.GetBizDB(ctx).Exec("CREATE DATABASE IF NOT EXISTS `" + databaseName + "`")
	if err != nil {
		return err
	}

	bizQuery, err := s.data.GetBizGormDBByName(oldDatabaseName)
	if err != nil {
		return err
	}
	db, err := bizQuery.DB()
	if err != nil {
		return err
	}
	defer db.Close()
	//// 锁表
	//if _, err = db.Exec("FLUSH TABLES WITH READ LOCK"); err != nil {
	//	return err
	//}
	//defer db.Exec("UNLOCK TABLES")
	// 查询所有的表
	var tables []string
	rows, err := db.QueryContext(ctx, "SHOW TABLES")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var table string
		if err = rows.Scan(&table); err != nil {
			return err
		}
		tables = append(tables, table)
	}

	for _, table := range tables {
		// 创建备份表
		err = bizQuery.WithContext(ctx).Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s LIKE %s", databaseName, table, table)).Error
		if err != nil {
			return err
		}
		// 备份数据
		err = bizQuery.WithContext(ctx).Exec(fmt.Sprintf("INSERT INTO %s.%s SELECT * FROM %s", databaseName, table, table)).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// resetTeam 删除数据库
func (s *systemRepositoryImpl) resetTeam(ctx context.Context, databaseName string) error {
	_, err := s.data.GetBizDB(ctx).Exec("DROP DATABASE IF EXISTS `" + databaseName + "`")
	return err
}

// restoreData 恢复团队数据
func (s *systemRepositoryImpl) restoreData(ctx context.Context, teamID uint32) error {
	databaseName := s.getBackupTeamDatabaseName(teamID)
	oldDatabaseName := data.GenBizDatabaseName(teamID)
	// 判断 databaseName 是否存在
	rows, err := s.data.GetBizDB(ctx).Query(fmt.Sprintf("SELECT count(SCHEMA_NAME) FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '%s'", databaseName))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var count int
		if err = rows.Scan(&count); err != nil {
			return err
		}
		if count == 0 {
			return nil
		}
	}
	// 备份数据库数据
	if err := s.backupTeam(ctx, oldDatabaseName, databaseName); err != nil {
		return err
	}
	return nil
}
