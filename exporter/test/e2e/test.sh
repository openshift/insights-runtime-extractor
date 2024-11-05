go test -p 1 -timeout 1h -count=1 -v 2>&1 ./test/e2e/... | tee test.log
${GOPATH}/bin/go-junit-report -in test.log  -set-exit-code > ${ARTIFACT_DIR:-.}/junit.xml
