//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/event"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/team"
	"github.com/spf13/cobra"
	"gorm.io/gen"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/system"
)

var c = gen.Config{
	OutPath: "./dal",
	Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	// 如果你希望为可为null的字段生成属性为指针类型, 设置 FieldNullable 为 true
	// FieldNullable: true,
	// 如果你希望在 `Create` API 中为字段分配默认值, 设置 FieldCoverable 为 true, 参考: https://gorm.io/docs/create.html#Default-Values
	FieldCoverable: true,
	// 如果你希望生成无符号整数类型字段, 设置 FieldSignable 为 true
	FieldSignable: true,
	// 如果你希望从数据库生成索引标记, 设置 FieldWithIndexTag 为 true
	FieldWithIndexTag: true,
	// 如果你希望从数据库生成类型标记, 设置 FieldWithTypeTag 为 true
	FieldWithTypeTag: true,
	// 如果你需要对查询代码进行单元测试, 设置 WithUnitTest 为 true
	// WithUnitTest: true,
}

var rootCmd = &cobra.Command{
	Use:   "gen",
	Short: "generate gorm code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome use generate gorm code")
		genSystem()
		genBiz()
		genAlarm()
	},
}

var sysCmd = &cobra.Command{
	Use:   "sys",
	Short: "generate system code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome use generate system code")
		genSystem()
	},
}

var bizCmd = &cobra.Command{
	Use:   "biz",
	Short: "generate biz code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome use generate biz code")
		genBiz()
	},
}

var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "generate event code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome use generate event code")
		genAlarm()
	},
}

func init() {
	rootCmd.AddCommand(sysCmd, bizCmd, eventCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func genSystem() {
	c.OutPath = "cmd/palace/internal/data/query/systemgen"
	g := gen.NewGenerator(c)
	g.ApplyBasic(system.Models()...)
	g.Execute()
}

func genBiz() {
	c.OutPath = "cmd/palace/internal/data/query/teamgen"
	g := gen.NewGenerator(c)
	g.ApplyBasic(team.Models()...)
	g.Execute()
}

func genAlarm() {
	c.OutPath = "cmd/palace/internal/data/query/eventgen"
	g := gen.NewGenerator(c)
	g.ApplyBasic(event.Models()...)
	g.Execute()
}
