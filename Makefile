root	:=		$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: clean build

clean:
	$(MAKE) -C "${root}/deploySecondStep" clean
	$(MAKE) -C "${root}/deployThirdStep" clean

build:
	$(MAKE) -C "${root}/deploySecondStep" build
	$(MAKE) -C "${root}/deployThirdStep" build
