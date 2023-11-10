package biz

import "github.com/google/wire"

// ProviderSetBiz is biz providers.
var ProviderSetBiz = wire.NewSet(NewPingUseCase)

type (
	IBO[B, D any] interface {
		DO() IDO[B, D]
		First() B
		List() []B
		DToB(D) B
		BToD(B) D
	}
	IDO[B, D any] interface {
		BO() IBO[B, D]
		First() D
		List() []D
		DToB(D) B
		BToD(B) D
	}

	BO[B, D any] struct {
		list []B
		bToD func(B) D
		dToB func(D) B
	}
	DO[B, D any] struct {
		list []D
		dToB func(D) B
		bToD func(B) D
	}

	BOOption[B, D any] func(*BO[B, D])
	DOOption[B, D any] func(*DO[B, D])
)

func (l *DO[B, D]) BO() IBO[B, D] {
	list := make([]B, 0, len(l.list))
	for _, v := range l.list {
		list = append(list, l.dToB(v))
	}
	return NewBO[B, D](BOWithValues[B, D](list...), BOWithDToB[B, D](l.dToB), BOWithBToD[B, D](l.bToD))
}

func (l *DO[B, D]) First() (one D) {
	if len(l.list) > 0 {
		one = l.list[0]
	}
	return
}

func (l *DO[B, D]) List() []D {
	return l.list
}

func (l *DO[B, D]) DToB(d D) B {
	return l.dToB(d)
}

func (l *DO[B, D]) BToD(b B) D {
	return l.bToD(b)
}

func (l *BO[B, D]) DO() IDO[B, D] {
	list := make([]D, 0, len(l.list))
	for _, v := range l.list {
		list = append(list, l.BToD(v))
	}
	return NewDO[B, D](DOWithValues[B, D](list...), DOWithDToB[B, D](l.dToB), DOWithBToD[B, D](l.bToD))
}

func (l *BO[B, D]) First() (one B) {
	if len(l.list) > 0 {
		one = l.list[0]
	}
	return
}

func (l *BO[B, D]) List() []B {
	return l.list
}

func (l *BO[B, D]) DToB(d D) B {
	return l.dToB(d)
}

func (l *BO[B, D]) BToD(b B) D {
	return l.bToD(b)
}

// NewBO new a biz.BO instance.
func NewBO[B, D any](opts ...BOOption[B, D]) IBO[B, D] {
	b := &BO[B, D]{
		list: make([]B, 0),
	}
	for _, opt := range opts {
		opt(b)
	}

	return b
}

// NewDO new a biz.DO instance.
func NewDO[B, D any](opts ...DOOption[B, D]) IDO[B, D] {
	d := &DO[B, D]{
		list: make([]D, 0),
	}
	for _, opt := range opts {
		opt(d)
	}

	return d
}

// BOWithBToD set bToD.
func BOWithBToD[B, D any](bToD func(B) D) BOOption[B, D] {
	return func(b *BO[B, D]) {
		b.bToD = bToD
	}
}

// BOWithDToB set dToB.
func BOWithDToB[B, D any](dToB func(D) B) BOOption[B, D] {
	return func(b *BO[B, D]) {
		b.dToB = dToB
	}
}

// BOWithValues set list.
func BOWithValues[B, D any](values ...B) BOOption[B, D] {
	return func(b *BO[B, D]) {
		b.list = append(b.list, values...)
	}
}

// DOWithValues set list.
func DOWithValues[B, D any](values ...D) DOOption[B, D] {
	return func(d *DO[B, D]) {
		d.list = append(d.list, values...)
	}
}

// DOWithDToB set dToB.
func DOWithDToB[B, D any](dToB func(D) B) DOOption[B, D] {
	return func(d *DO[B, D]) {
		d.dToB = dToB
	}
}

// DOWithBToD set bToD.
func DOWithBToD[B, D any](bToD func(B) D) DOOption[B, D] {
	return func(d *DO[B, D]) {
		d.bToD = bToD
	}
}
