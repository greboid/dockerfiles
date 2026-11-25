.PHONY: test vet check yaml images

test:
	go test ./...

vet:
	go vet ./...

check: test vet yaml

# action.yml files are parsed by the runner, not by go — validate them
# when python3 with yaml is available (silently skipped otherwise).
yaml:
	@python3 -c 'import yaml' 2>/dev/null || { echo "python3 yaml module unavailable; skipping action.yml validation"; exit 0; }
	for f in */action.yml; do python3 -c 'import yaml,sys; yaml.safe_load(open(sys.argv[1])); print(sys.argv[1] + ": valid yaml")' $$f; done

# Build every docker action image locally, as the images workflow does before
# pushing them: the docker build context is the repository root, not the
# action directory. Tags match the registry refs the workflow pushes.
images:
	for a in git-checkout go-setup node-setup apko-changelog \
		docker-login docker-build docker-push \
		go-build go-test go-vet go-lint \
		git-commit git-try-push release package-latest; do \
		docker build -f Dockerfile.$$a -t git.mouse-lake.ts.net/actions/toolkit/$$a:dev . || exit 1; \
	done
