FROM golang
RUN dpkg --add-architecture i386 && apt-get update
RUN apt-get install -y --no-install-recommends cpp:i386 gcc:i386 g++:i386 autotools-dev:i386 autoconf:i386 automake:i386 libudev-dev:i386 libusb-1.0-0-dev:i386
RUN git clone https://github.com/signal11/hidapi
RUN apt-get install -y libtool:i386
RUN cd hidapi && ./bootstrap && ./configure CFLAGS="-m32" && make CFLAGS="-m32" && make install
CMD cd /go/src/betcoin && make binary APP=$APP
