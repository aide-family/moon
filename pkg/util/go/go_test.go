package goroutine_test

import (
	"context"
	"fmt"
	"time"

	goroutine "github.com/aide-family/moon/pkg/util/go"
)

func ExampleNew() {
	limit := 2
	multiple := 2
	g := goroutine.New(limit, multiple)
	ctx := context.Background()
	if err := g.Start(ctx); err != nil {
		panic(err)
	}

	for i := range 20 {
		item := i
		goroutine.Go(func() {
			time.Sleep(100 * time.Millisecond)
			fmt.Println("msg: ", item)
		})
	}

	time.Sleep(time.Second * 4)

	if err := g.Stop(ctx); err != nil {
		panic(err)
	}
}
