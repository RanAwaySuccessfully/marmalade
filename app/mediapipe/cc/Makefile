#CFLAGS = -c -fPIC
CPPFLAGS=-I./mediapipe/
LDFLAGS=-L./ -lmediapipe

#OBJS = libtoast-bodym.o libtoast-facem.o libtoast-handm.o libtoast.o
OBJS = libtoast-posem.o libtoast-facem.o libtoast-handm.o libtoast.o

OPTS = -O2

all: libtoast.a clean-o

libtoast.a : $(OBJS)
	$(AR) rcs $@ $^

.c.o:
	$(CC) $(CFLAGS) $(OPTS) $*.cc 

clean-o:
	rm -f $(OBJS) > /dev/null

clean:
	rm -f $(OBJS) libtoast.a > /dev/null

