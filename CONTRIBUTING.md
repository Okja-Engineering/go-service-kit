# Contributing to go-service-kit

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are greatly appreciated!

## House Rules for PRs and Issues

### 👥 Prevent Work Duplication
Before submitting a new issue or PR, check if it already exists in the Issues or Pull Requests.

### ✅ Work Only on Approved Issues
- For feature requests, please wait for a core team member to approve and remove the 🚨 needs approval label before you start coding or submitting a PR.
- For bugs, security, performance, or documentation, you can start coding immediately—even if the 🚨 needs approval label is present.

### 🚫 Don’t Just Drop a Link
Avoid posting third-party links (e.g., Slack threads) without context. Every GitHub issue or PR should stand on its own.

### 👀 Think Like a Reviewer
Put yourself in the reviewer’s shoes. What would you want to know if reading this for the first time? Are there key decisions, goals, or constraints that need clarification? Does the PR assume knowledge that isn’t obvious? Are there related issues or previous PRs that should be linked?

### 🧵 Bring in Context from Private Channels
If the task originated from a private conversation, extract the relevant details and include them in the GitHub issue or PR (avoid sharing sensitive info).

### 📚 Treat It Like Documentation
Write clearly enough that someone—possibly you—can revisit it months later and still understand what happened and why.

### ✅ Summarize Your PR at the Top
Even if the code changes are minor, a short written summary helps reviewers quickly understand the intent.

### 🔗 Use GitHub Keywords to Auto-Link Issues
Use phrases like “Closes #123” or “Fixes #456” in your PR descriptions.

### 🧪 Mention What Was Tested (and How)
Explain how you validated your changes. Example:  
“Tested locally with mock data and confirmed the flow works.”

### 🧠 Assume Future-You Won’t Remember
If there are trade-offs, edge cases, or temporary workarounds, document them clearly.

## Priorities

| Type of Issue                        | Priority         |
|-------------------------------------- |-----------------|
| Minor improvements, non-core features | Low             |
| Confusing DX (but still functional)   | Medium          |
| Core Features (API, auth, env, etc.)  | High            |
| Core Bugs (build, test, lint fails)   | Highest         |

## File Naming Conventions

- Test files: `*_test.go`
- Package files: Use clear, descriptive names matching the main exported type or function.
- Avoid ambiguous names like `util.go` or `misc.go`.

## Developing

See the [README](./README.md) for setup, building, and testing instructions.

## Building

This repo is a set of Go packages, not an executable. You can validate changes by running:

```sh
make test
make lint
```

## Testing

Tests are located alongside each package. Add or update tests as needed to cover your changes.

## Making a Pull Request

- Check the "Allow edits from maintainers" option when creating your PR.
- If your PR refers to or fixes an issue, add `refs #XXX` or `fixes #XXX` to the PR description.
- Fill out the PR template accordingly.
- Keep your branches updated (e.g., click the Update branch button on the GitHub PR page).
