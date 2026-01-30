# Contributing to go-report-engine

Thank you for your interest in contributing to **go-report-engine**! We welcome contributions from the community to make this project better.

## Ways to Contribute

1.  **Report Bugs**: Open an issue describing the bug, including steps to reproduce and expected behavior.
2.  **Suggest Features**: Open an issue or discussion to propose new features or improvements.
3.  **Submit Pull Requests**: Implement bug fixes or new features and submit a PR for review.
4.  **Improve Documentation**: Help us make the documentation clearer and more comprehensive.

## Development Workflow

1.  **Fork the Repository**: specific to your GitHub account.
2.  **Clone the Fork**: `git clone https://github.com/YourUsername/go-report-engine.git`
3.  **Create a Branch**: `git checkout -b feature/my-new-feature` or `git checkout -b fix/bug-fix`
4.  **Make Changes**: Write clean, testable code following the project's style.
5.  **Run Tests**: Ensure all tests pass.
    ```bash
    go test ./... -v
    go test ./... -race
    ```
6.  **Run Linters**:
    ```bash
    go vet ./...
    # If installed: golangci-lint run
    ```
7.  **Commit Changes**: Write clear, descriptive commit messages.
8.  **Push to Fork**: `git push origin my-branch`
9.  **Open Pull Request**: detailed description of your changes.

## Coding Guidelines

-   **Go Style**: Follow standard Go conventions (Effective Go).
-   **SOLID Principles**: Adhere to SOLID principles for a modular and maintainable architecture.
-   **Documentation**:
    -   Exported functions, types, and constants **MUST** have godoc comments.
    -   Update `README.md` or other docs if you change behavior.
-   **Testing**:
    -   New code must have unit tests.
    -   Maintain or improve code coverage (aim for >95%).
    -   Use table-driven tests where appropriate.
    -   Avoid flaky tests.

## Code Review Process

-   All PRs require review by a maintainer.
-   Address feedback constructively.
-   Once approved, your PR will be merged.

## Community

-   Be respectful and kind to others.
-   Follow the [Code of Conduct](CODE_OF_CONDUCT.md).

Happy coding! 🚀
