install:
	go install .

release:
	goreleaser --rm-dist
