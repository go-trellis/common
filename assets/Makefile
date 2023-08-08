.PHONY: testdata
testdata:
	find testdata -type f -name '*.gz' -delete
	find testdata -type f -exec sh -c 'gzip -fc {} > {}.gz' \; -print0 | xargs -0 -I % echo %.gz
	mv testdata/uncompressed.gz testdata/compressed.gz

.PHONY: test
test:
	go test -v ./...