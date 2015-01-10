CC=gcc
CFLAGS=-Wall -Wextra -pedantic -std=c99 -O0
TESTFILE=known_byte_sequences.c

all: test test32

test:
	$(CC) $(CFLAGS) $(TESTFILE) -o test

test32:
	$(CC) $(CFLAGS) $(TESTFILE) -m32 -o test

clean:
	rm test tests32
