package system_test

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/plugin/gorm"
	"github.com/aide-family/moon/pkg/util/password"
	"github.com/joho/godotenv"
)

var c config.Database

func init() {
	_ = godotenv.Load(".env")
	port, err := strconv.Atoi(os.Getenv("X_MOON_DATA_MAIN_PORT"))
	if err != nil {
		return
	}
	c = config.Database{
		Driver:       config.Database_Driver(config.Database_Driver_value[os.Getenv("X_MOON_DATA_MAIN_DRIVER")]),
		User:         os.Getenv("X_MOON_DATA_MAIN_USER"),
		Password:     os.Getenv("X_MOON_DATA_MAIN_PASSWORD"),
		Host:         os.Getenv("X_MOON_DATA_MAIN_HOST"),
		Port:         int32(port),
		Params:       os.Getenv("X_MOON_DATA_MAIN_PARAMS"),
		Debug:        true,
		UseSystemLog: false,
		DbName:       os.Getenv("X_MOON_DATA_MAIN_DB_NAME"),
	}
}

func Test_NewUser(t *testing.T) {
	db, err := gorm.NewDB(&c)
	if err != nil {
		bs, _ := json.Marshal(&c)
		t.Fatalf("err: %s, config: %s", err, string(bs))
		return
	}

	pass := password.New("123456")
	enPass, err := pass.EnValue()
	if err != nil {
		t.Fatal(err)
		return
	}

	user := &system.User{
		Username: "admin",
		Nickname: "管理员",
		Password: enPass,
		Email:    "1058165620@qq.com",
		Phone:    "",
		Remark:   "",
		Avatar:   "",
		Salt:     pass.Salt(),
		Gender:   vobj.GenderMale,
		Position: vobj.RoleSuperAdmin,
		Status:   vobj.UserStatusNormal,
		Roles:    nil,
		Teams:    nil,
	}

	if err := db.GetDB().Create(user).Error; err != nil {
		t.Fatal(err)
		return
	}
}
