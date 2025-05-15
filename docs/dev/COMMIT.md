<div style="display: flex; align-items: center;">
  <img 
    src="../images/logo.svg" 
    alt="Logo" 
    style="height: 4em; width: auto; vertical-align: middle; margin-right: 10px;" 
  />
  <h1 style="margin: 0; font-size: 24px; line-height: 1.5;">Moon - Multi-Domain Monitoring and Alerting Platform</h1>
</div>

| [English](COMMIT.md) | [简体中文](COMMIT.zh-CN.md) |

## 1. Commit Message Format

Each commit message should follow the format below:

```bash
<type>(<scope>): <subject>
<blank line>
<body>
<blank line>
<footer>
```

### 1.1 Type

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation changes
- **style**: Code style changes (does not affect code execution)
- **refactor**: Code refactoring (neither fixes a bug nor adds a feature)
- **perf**: Performance improvements
- **test**: Adding or modifying tests
- **chore**: Changes to the build process or auxiliary tools
- **revert**: Reverting to a previous commit

### 1.2 Scope

- Optional field, indicating the scope or module affected. For example: user, auth, api, config, etc.

### 1.3 Subject

- A brief description, no more than 50 characters.
- Use imperative mood, lowercase first letter, and no period at the end.

### 1.4 Body

- Optional field, providing a detailed description of the changes.
- Use imperative mood, with each line no more than 72 characters.
- Explain why the change was made and how it was implemented.

### 1.5 Footer

- Optional field, used to reference issues or close them.
- For example: Closes #123, Fixes #456.

## 2. Commit Message Examples

### 2.1 New Feature

```bash
feat(user): Add user registration feature

- Implement user registration functionality
- Add relevant test cases

Closes #123
```

### 2.2 Bug Fix

```bash
fix(auth): Fix login failure issue

- Fix password verification failure during login
- Add relevant test cases

Fixes #456
```

### 2.3 Documentation Update

```bash
docs(readme): Update project README file

- Add project installation steps
- Update project dependency instructions
```

### 2.4 Code Refactoring

```bash
refactor(api): Refactor user API interface

- Split user API interface into multiple modules
- Optimize code structure for better readability
```

## 3. Commit Frequency

- Strive for atomic commits, where each commit contains only one feature or fix.
- Avoid submitting large amounts of code at once; ensure each commit has a clear purpose.

## 4. Branch Management

- main: The main branch, used for releasing stable versions.
- develop: The development branch, used for integrating new features.
- feature/xxx: Feature branches, used for developing new features.
- bugfix/xxx: Bug fix branches, used for fixing issues.
- hotfix/xxx: Hotfix branches, used for urgent fixes in the production environment.

## 5. Pull Request

- After completing a feature or fix, create a pull request to the `develop` branch.
- The pull request should include a detailed description explaining the changes and their purpose.
- The pull request must pass code review before merging.

## 6. Code Review

- Ensure all tests pass before submitting code.
- After submission, team members should conduct a code review to ensure code quality and consistency.

## 7. Other Considerations

- Avoid committing sensitive information (e.g., passwords, keys).
- Ensure code is formatted according to the team's code style guide before committing.

---

**Note**: This specification aims to improve team collaboration efficiency and code quality. Please adhere to it strictly.