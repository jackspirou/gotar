build:
	go install

release:
	go get github.com/mitchellh/gox
	gox
	tarpack tarpack_*
