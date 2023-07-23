package main

import (
	"fmt"
	"prometheus-manager/apps/node/internal/conf"
)

func fmtASCIIGenerator(env *conf.Env) {
	fmt.Println(Name + " service starting...")

	fmt.Println(`┌───────────────────────────────────────────────────────────────────────────────────────┐
│                                      _____  _____   ______                            │
│                               /\    |_   _||  __ \ |  ____|                           │
│                              /  \     | | || |  | || |__                              │
│                             / /\ \    | | || |  | ||  __|                             │
│                            / /__\ \  _| |_|| |__| || |____                            │
│                           /_/    \_\|_____||_____/ |______|                           │							
│                                 good luck and no bug                                  │
└───────────────────────────────────────────────────────────────────────────────────────┘`)

	detail := `
┌──────────────────────────────────────────────────────────────────────────────────────`

	detail += fmt.Sprintf("\n├── %s: %s", "Name", Name)
	detail += fmt.Sprintf("\n├── %s: %s", "Version", Version)
	detail += fmt.Sprintf("\n├── %s: %s", "ID", id)
	detail += fmt.Sprintf("\n├── %s: %s", "Metadata", "")
	for k, p := range env.Metadata {
		detail += fmt.Sprintf("\n├────── %s: %s", k, p)
	}

	detail += `
└──────────────────────────────────────────────────────────────────────────────────────
`

	fmt.Println(detail)
}
