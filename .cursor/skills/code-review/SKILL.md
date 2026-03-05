---
name: code-review
description: Reviews code for correctness and potential bugs, pinpoints bug locations by file and line, and suggests concrete fixes. Use when the user asks for a code review, wants to find bugs, or mentions reviewing code or changes.
---

# Code Review

## Objective

Review the given code for **correctness and potential bugs**. For each issue found, report the **exact location (file and line)** and provide a **recommended fix**.

## Review Process

1. **Read the code** in scope (file(s) or diff the user provided).
2. **Identify potential bugs**, including:
   - Logic errors, off-by-one, wrong conditions
   - Nil/zero dereference, out-of-bounds access
   - Race conditions, concurrency misuse
   - Error handling gaps (ignored errors, wrong propagation)
   - Resource leaks (unclosed handles, connections)
   - Incorrect assumptions about types, APIs, or data
3. **For each finding**: state file path, line number(s), and a short recommended solution.

## Output Format

Use this structure for the review:

```markdown
# Code Review

## Summary
[1–2 sentence overview: scope and overall risk level]

## Findings

### [Severity] Brief title
- **Location**: `path/to/file.ext:LINE` (and optional `:END_LINE` if spanning multiple lines)
- **Issue**: What is wrong and why it can cause a bug.
- **Recommendation**: Concrete fix (code snippet or step-by-step change).

---
[Repeat for each finding]
```

**Severity**:
- **Critical**: Likely crash, data corruption, or security impact.
- **High**: Clear bug under common usage.
- **Medium**: Edge-case bug or maintainability/robustness concern.
- **Low**: Minor or speculative; worth fixing when touching the area.

## Rules

- **Always cite location**: Use `path:line` (and line range if needed). Do not say "somewhere in this function" without the line.
- **One finding per item**: One bug/issue per subsection; do not merge unrelated issues.
- **Recommendation must be actionable**: Prefer a small code change or clear steps, not vague advice.
- **Assume context**: If the project has patterns (e.g. error handling, logging), align recommendations with them; do not invent new patterns without need.

## Example

```markdown
### [High] Possible nil pointer dereference
- **Location**: `app/service/user.go:42`
- **Issue**: `u` may be nil when `FindByID` returns no user; calling `u.Name` can panic.
- **Recommendation**: Check and handle nil before use:

  if u == nil {
      return nil, ErrUserNotFound
  }
  return u.Name, nil
```
