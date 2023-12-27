package hello

import (
	"fmt"
	"os"
)

var (
	name     = ""
	version  = ""
	env      = ""
	metadata = map[string]string{}
	id, _    = os.Hostname()
)

// SetName set name
func SetName(n string) {
	name = n
}

func SetVersion(v string) {
	version = v
}

func SetEnv(e string) {
	env = e
}

func SetMetadata(m map[string]string) {
	metadata = m
}

// Name name
func Name() string {
	return name
}

func Version() string {
	return version
}

func Env() string {
	return env
}

func Metadata() map[string]string {
	return metadata
}

func ID() string {
	return id
}

func FmtASCIIGenerator() {
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

	detail := `
┌───────────────────────────────────────────────────────────────────────────────────────`

	detail += fmt.Sprintf("\n├── %s: %s", "Name", name)
	detail += fmt.Sprintf("\n├── %s: %s", "Version", version)
	detail += fmt.Sprintf("\n├── %s: %s", "ID", id)
	detail += fmt.Sprintf("\n├── %s: %s", "Env", env)
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
