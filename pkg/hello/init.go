package hello

import (
	"fmt"
	"os"
)

func FmtASCIIGenerator(name, version string, metadata map[string]string) {
	fmt.Println(name + " service starting...")

	fmt.Println(`┌───────────────────────────────────────────────────────────────────────────────────────┐
│                                      _____  _____   ______                            │
│                               /\    |_   _||  __ \ |  ____|                           │
│                              /  \     | | || |  | || |__                              │
│                             / /\ \    | | || |  | ||  __|                             │
│                            / /__\ \  _| |_|| |__| || |____                            │
│                           /_/    \_\|_____||_____/ |______|                           │							
│                                 good luck and no bug                                  │
└───────────────────────────────────────────────────────────────────────────────────────┘`)

	id, _ := os.Hostname()

	detail := `
┌───────────────────────────────────────────────────────────────────────────────────────`

	detail += fmt.Sprintf("\n├── %s: %s", "Name", name)
	detail += fmt.Sprintf("\n├── %s: %s", "Version", version)
	detail += fmt.Sprintf("\n├── %s: %s", "ID", id)
	if metadata != nil {
		detail += fmt.Sprintf("\n├── %s: %s", "Metadata", "")
		for k, p := range metadata {
			detail += fmt.Sprintf("\n├────── %s: %s", k, p)
		}
	}

	detail += `
└───────────────────────────────────────────────────────────────────────────────────────
`

	fmt.Println(detail)
}
