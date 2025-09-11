Contributions are welcomed. Please follow these guidelines:

## Getting Started

1. Fork the repository
2. Clone your fork with submodules: `git clone --recursive <your-fork-url>`
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Test your changes thoroughly
6. Submit a pull request

## Development Guidelines

- **Code Style**: Follow standard Go formatting (`go fmt`)
- **Testing**: All new features must include tests
- **Documentation**: Update documentation for any API changes
- **Commits**: Use clear, descriptive commit messages

## Testing

If you've already cloned the repository without the submodule, initialize it:

```bash
git submodule update --init --recursive
```

To update the test suite submodule to the latest version:

```bash
git submodule update --remote
```

### Prerequisites

For comprehensive testing, this project uses xdelta3 as a reference implementation to verify the correctness of the decoder.

#### Installing xdelta3

##### macOS (Homebrew)
```bash
brew install xdelta
```

##### Linux (Ubuntu/Debian)
```bash
sudo apt-get install xdelta3
```

### Running Tests

To run the Go unit tests:

```bash
go test ./...
```

To run the comprehensive test suite against the VCDIFF test cases (requires submodule):

```bash
cd submodules/vcdiff-tests
./run_tests.sh ../../vcdiff
```

To run with coverage analysis:

```bash
./coverage.sh
```

The test suite includes:
- **57 positive tests**: Valid VCDIFF files that should decode successfully
- **37 negative tests**: Invalid VCDIFF files that should be rejected with appropriate errors
- **Fuzz testing**: `go test -fuzz=.` for robustness testing


## Reporting Issues

When reporting bugs, please include:
- Go version
- Operating system
- Minimal reproduction case
- Expected vs actual behavior
- Sample VCDIFF files (if applicable)

## Feature Requests

For new features, please:
- Check existing issues first
- Describe the use case
- Provide RFC 3284 references if applicable
- Consider backwards compatibility

## Release process
1. Create a branch for the release, named like `release/1.2.3` (where `1.2.3` is the new version number)
2. Run [`github_changelog_generator`](https://github.com/github-changelog-generator/github-changelog-generator) to automate the update of the [CHANGELOG](./CHANGELOG.md). This may require some manual intervention, both in terms of how the command is run and how the change log file is modified. Your mileage may vary:
  - The command you will need to run will look something like this: `github_changelog_generator -u ably -p vcdiff-go --since-tag v1.2.3 --output delta.md --token $GITHUB_TOKEN_WITH_REPO_ACCESS`. Generate token [here](https://github.com/settings/tokens/new?description=GitHub%20Changelog%20Generator%20token).
  - Using the command above, `--output delta.md` writes changes made after `--since-tag` to a new file
  - The contents of that new file (`delta.md`) then need to be manually inserted at the top of the `CHANGELOG.md`, changing the "Unreleased" heading and linking with the current version numbers
  - Also ensure that the "Full Changelog" link points to the new version tag instead of the `HEAD`
3. Commit this change: `git add CHANGELOG.md && git commit -m "Update change log."`
4. Make a PR against `main`
5. Once the PR is approved, merge it into `main`
6. Add a tag to the new `main` head commit and push to origin such as `git tag v1.2.3 && git push origin v1.2.3`
7. Create the release on Github, from the new tag, including populating the release notes
8. Create the entry on the [Ably Changelog](https://changelog.ably.com/) (via [headwayapp](https://headwayapp.co/))