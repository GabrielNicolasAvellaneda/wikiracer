build:
	@echo "+$@"
	go build
install:
	@echo "+$@"
	go install
test:
	@echo "+$@"
	bash ./script/test.sh
clean:
	@echo "+$@"
	rm -f ./wikiracer
build-image:
	@echo "+$@"
	bash ./script/build.sh
run:
	@echo "+$@"
	bash ./script/run.sh
