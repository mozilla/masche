TESTBINDIR=test/tools
TESTS=./memaccess ./memsearch ./process

all: run_tests run_tests32

run_tests: testbin
	go test $(TESTS)

testbin:
	$(MAKE) -C $(TESTBINDIR) test

run_tests32: testbin32
	go test $(TESTS)

testbin32:
	$(MAKE) -C $(TESTBINDIR) test32
