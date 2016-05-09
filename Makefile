ginkgo_cli :
		go install github.com/onsi/ginkgo/ginkgo

build :
		go build -v ./...

test : ginkgo_cli build
		go get -v -t ./...
		ginkgo -v -r -randomizeSuites -randomizeAllSpecs -race
