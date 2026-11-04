---
description: "Use when writing or editing Go code. Blank-line and spacing style: separate multi-line statements, blocks and logical groups with empty lines."
applyTo: "**/*.go"
---
# Go style: blank lines

Blank lines show the steps of a function. Code that only passes `gofmt` usually feels too dense
here. `gofmt` keeps blank lines but never adds them, so add them yourself.

1. **After a statement that spans multiple lines** (a long call, a composite literal, a func
   literal), add a blank line.
2. **After the closing `}` of a block** (`if`, `for`, `switch`, `select`, func literal), add a blank
   line, unless the next line is another `}`.
3. **Between logical groups** in a function (setup, doing the work, handling the result), add a
   blank line. Keep closely related one-liners together, e.g. a run of field assignments, or
   `x, err := f()` directly followed by `if err != nil`.

```go
// Bad
cmd := exec.CommandContext(ctx, "copilot",
	"-p", prompt,
)
cmd.Dir = dir
cmd.Stdout = out
if err := cmd.Run(); err != nil {
	return fmt.Errorf("run copilot: %w", err)
}
return nil

// Good
cmd := exec.CommandContext(ctx, "copilot",
	"-p", prompt,
)

cmd.Dir = dir
cmd.Stdout = out

if err := cmd.Run(); err != nil {
	return fmt.Errorf("run copilot: %w", err)
}

return nil
```

**Check:** run `golangci-lint run ./...` from `platform/` after editing Go code. It enforces rule 2
(`wsl_v5` `after-block`, configured in `platform/.golangci.yml`), and `--fix` adds the missing lines.
No tool checks rules 1 and 3, so follow them yourself.
