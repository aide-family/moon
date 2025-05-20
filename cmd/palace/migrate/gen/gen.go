//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do/event"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/spf13/cobra"
	"gorm.io/gen"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
)

var c = gen.Config{
	OutPath: "./dal",
	Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	// If you want to generate pointer type properties for nullable fields, set FieldNullable to true
	// FieldNullable: true,
	// If you want to assign default values to fields in the `Create` API, set FieldCoverable to true, see: https://gorm.io/docs/create.html#Default-Values
	FieldCoverable: true,
	// If you want to generate unsigned integer type fields, set FieldSignable to true
	FieldSignable: true,
	// If you want to generate index tags from the database, set FieldWithIndexTag to true
	FieldWithIndexTag: true,
	// If you want to generate type tags from the database, set FieldWithTypeTag to true
	FieldWithTypeTag: true,
	// If you need unit tests for query code, set WithUnitTest to true
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
