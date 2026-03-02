# Contributing to LuckyClaw

LuckyClaw is currently maintained by a single developer. Contributions via pull requests and issue reports are welcome on GitHub. Response times may vary.

## Before Submitting a PR

1. Run `make check` and ensure all tests pass locally.
2. Keep PRs focused. Avoid bundling unrelated changes together.
3. Include a clear description of what changed and why.

## PR Structure

Every pull request should include:

- **Description** -- What does this change do and why?
- **Type** -- Bug fix, feature, docs, or refactor.
- **Testing** -- How you tested the change (hardware, model/provider, channel).
- **Evidence** -- Logs or screenshots demonstrating the change works (optional but encouraged).

## AI-Assisted Contributions

LuckyClaw embraces AI-assisted development. If you use AI tools to generate code, please:

- **Disclose it** in the PR description. There is no stigma -- only transparency matters.
- **Read and understand** every line of generated code before submitting.
- **Test it** in a real environment, not just in an editor.
- **Review for security** -- AI-generated code can produce subtle bugs around path traversal, command injection, and credential handling.

AI-generated contributions are held to the same quality bar as human-written code.

## Code Standards

- Idiomatic Go, consistent with the existing codebase style.
- No unnecessary abstractions, dead code, or over-engineering.
- Include or update tests where appropriate.
- All CI checks (`make check`) must pass.

## Communication

- **GitHub Issues** -- Bug reports, feature requests, design discussions.
- **Pull Request comments** -- Code-specific feedback.

When in doubt, open an issue before writing code.
