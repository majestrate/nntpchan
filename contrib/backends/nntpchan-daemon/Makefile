
EXE = nntpd

TOOL = nntpchan-tool

CXX = g++
REPO=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SRC_PATH = $(REPO)/src

SOURCES := $(wildcard $(SRC_PATH)/*.cpp)
HEADERS := $(wildcard $(SRC_PATH)/*.hpp)
OBJECTS := $(SOURCES:.cpp=.o)

PKGS := libuv libsodium

LD_FLAGS := $(shell pkg-config --libs $(PKGS))
INC_FLAGS := $(shell pkg-config --cflags $(PKGS)) -I $(REPO)/src
CXXFLAGS := -std=c++11 -Wall -Wextra $(INC_FLAGS) -g

LIB = libnntpchan.a

all: $(EXE) $(TOOL)

$(LIB): $(OBJECTS)
	$(AR) -r $(LIB) $(OBJECTS)

$(EXE): $(LIB)
	$(CXX) -o $(EXE) $(CXXFLAGS) nntpd.cpp $(LIB) $(LD_FLAGS)

$(TOOL): $(LIB)
	$(CXX) -o $(TOOL) $(CXXFLAGS) tool.cpp $(LIB) $(LD_FLAGS)

build-test: $(LIB)
	$(CXX) -o test $(CXXFLAGS) test.cpp $(LIB) $(LD_FLAGS)

test: build-test
	./test

%.o: src/%.cpp
	$(CXX) $(CXXFLAGS) -c -o $@

clean:
	rm -f $(OBJECTS) $(LIB) $(EXE) $(TOOL) test
