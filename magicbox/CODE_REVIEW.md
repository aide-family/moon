# Code Review

## Summary

Review of the `magicbox` package for correctness and potential bugs. Scope: server/middler (validate), oauth/feishu, strutil/encrypt, safety (go, map, slices), and related tests. **Overall risk: medium** — several clear bugs (wrong error API, unused option, nil map/slice construction) and a few robustness items.

---

## Findings

### [High] Feishu: Wrong first argument to errors.New (API code used as HTTP status)

- **Location**: `magicbox/oauth/feishu/feishu.go:41`
- **Issue**: `errors.New(userResponse.Code, "GET_USER_INFO_FAILED", userResponse.Msg)` uses Feishu’s API response code as the first parameter. Kratos `errors.New(code, reason, message)` expects an **HTTP status code** (e.g. 200–599). Feishu codes (e.g. 40001, 99991663) are not valid HTTP statuses and can break error handling / mapping to HTTP/gRPC.
- **Recommendation**: Use a fixed HTTP status (e.g. 500) and put Feishu code and message in reason/message or metadata:

```go
return nil, merr.ErrorInternalServer("get user info failed").
    WithCause(errors.New(500, "GET_USER_INFO_FAILED", userResponse.Msg)).
    WithMetadata(map[string]string{"feishu_code": strconv.Itoa(userResponse.Code)})
```

Or at minimum:

```go
return nil, merr.ErrorInternalServer("get user info failed").WithCause(
    errors.New(500, "GET_USER_INFO_FAILED", userResponse.Msg))
```

---

### [High] Validate: protovalidate option never applied

- **Location**: `magicbox/server/middler/validate.go:70`
- **Issue**: `protovalidate.WithMessages()` is called but its return value is discarded. Validator options must be passed to `protovalidate.New(opts...)`. As written, this call has no effect.
- **Recommendation**: Either pass the option into `New`, or remove the line if no messages are intended:

```go
func validateParams(opts ...protovalidate.ValidatorOption) ValidateHandler {
	opts = append(opts, protovalidate.WithMessages()) // if you need WithMessages
	validator, err := protovalidate.New(opts...)
	if err != nil {
		panic(err)
	}
	// remove the standalone: protovalidate.WithMessages()
	return func(ctx context.Context, req interface{}) error {
		// ...
	}
}
```

If `WithMessages()` was meant to take message types, pass them and include in `opts` before calling `New`.

---

### [High] safety.Map: NewMap(nil) produces map that panics on write

- **Location**: `magicbox/safety/map.go:46–47`
- **Issue**: `maps.Clone(nil)` returns `nil`, so `NewMap(nil)` yields `&Map{K,V]{m: nil}`. Reads (e.g. `Get`, `Len`) are fine, but any write (`Set`, `Append`, `Delete`, `Clear`, etc.) does operations on a nil map and **panics**.
- **Recommendation**: Treat nil input as “empty map” so the internal map is never nil:

```go
func NewMap[K comparable, V any](m map[K]V) *Map[K, V] {
	if m == nil {
		return &Map[K, V]{m: make(map[K]V)}
	}
	return &Map[K, V]{m: maps.Clone(m)}
}
```

---

### [Medium] safety.Slice: NewSlice(nil) and missing bounds checks

- **Location**: `magicbox/safety/slices.go:23–25` and `magicbox/safety/slices.go:27–37`
- **Issue**: (1) `NewSlice(nil)` gives `&Slice{T]{s: nil}`; `Get(i)` / `Set(i, v)` then use `s.s[i]` and can panic (index out of range or write to nil). (2) `Get(i)` and `Set(i, v)` do not check `0 <= i < len(s.s)`, so out-of-range index causes panic.
- **Recommendation**: (1) In `NewSlice`, if `s == nil` set internal slice to `nil` or `make([]T, 0)` and document behavior; if you want nil-safe writes, use a non-nil slice (e.g. `slices.Clone(s)` only when `s != nil`, else `[]T{}`). (2) Add bounds checks in `Get`/`Set` (or document that callers must ensure valid index):

```go
func NewSlice[T any](s []T) *Slice[T] {
	if s == nil {
		return &Slice[T]{s: []T{}}
	}
	return &Slice[T]{s: slices.Clone(s)}
}

func (s *Slice[T]) Get(i int) T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if i < 0 || i >= len(s.s) {
		var zero T
		return zero
	}
	return s.s[i]
}
```

(Similarly for `Set` to avoid panic or return an error/bool.)

---

### [Medium] strutil.EncryptString: global encrypt can be nil after SetEncrypt(nil)

- **Location**: `magicbox/strutil/encrypt.go:36–38` and `magicbox/strutil/encrypt.go:55–56`
- **Issue**: `Value()` and `Scan()` use the package-level `encrypt`. If the first call to `SetEncrypt` is `SetEncrypt(nil)`, then `encrypt` stays nil and `encrypt.Encrypt` / `encrypt.Decrypt` will panic.
- **Recommendation**: Guard against nil in `Value()` and in `Scan()` before calling `encrypt.Decrypt`:

```go
func (e EncryptString) Value() (driver.Value, error) {
	if encrypt == nil {
		return "", nil
	}
	return encrypt.Encrypt(string(e))
}

// In Scan, before calling encrypt.Decrypt:
if encrypt == nil {
	return fmt.Errorf("encrypt not configured")
}
```

Alternatively, forbid nil in `SetEncrypt`:

```go
func SetEncrypt(e EncryptInterface) {
	once.Do(func() {
		if e != nil {
			encrypt = e
		}
	})
}
```

---

### [Low] safety.Map: Value() on nil receiver

- **Location**: `magicbox/safety/map.go:26–28`
- **Issue**: If someone calls `(*Map[K,V]).Value()` on a nil receiver (e.g. `var m *safety.Map[string,int]` then `m.Value()`), the method will dereference `m` to access `m.m` and panic.
- **Recommendation**: Either document that `*Map` must not be nil when calling methods, or add a nil check at the start of `Value()` (and other methods that touch `m.m`) and return a sensible default (e.g. `json.Marshal(nil)` → `[]byte("null"), nil`).

---

### [Low] validate.go: variable shadowing of `err`

- **Location**: `magicbox/server/middler/validate.go:76`
- **Issue**: The closure reuses the outer `err` from `validateParams` (line 65). The line `err = validator.Validate(message)` overwrites that. This is valid but can make debugging harder; linters often flag assignment to outer `err` in closures.
- **Recommendation**: Use a new variable inside the closure, e.g. `validateErr := validator.Validate(message)` and then use `validateErr` in the rest of the closure, to avoid shadowing and make control flow clearer.

---

## Positive notes

- **safety.Go**: Panic is recovered and logged; context is passed through correctly.
- **safety.Map**: Thread-safe and implements `Valuer`/`Scanner` with nil and type handling in `Scan`.
- **strutil.EncryptString**: Handles nil and string/[]byte in `Scan`; tests cover round-trip and invalid types.
- **validate**: Good use of `errors.As` and structured error with metadata for validation violations.
