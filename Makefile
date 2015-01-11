TESTBINDIR=test/tools
TESTS=./memaccess ./memsearch ./process

all: run_tests64 run_tests32

run_tests64: testbin64
	go test $(TESTS)

testbin64:
	$(MAKE) -C $(TESTBINDIR) test64

run_tests32: testbin32
	go test $(TESTS)

testbin32:
	$(MAKE) -C $(TESTBINDIR) test32

clean:
	go clean $(TESTS)
	$(MAKE) -C $(TESTBINDIR) clean
