
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



all: $(EXE) $(TOOL)

$(EXE): $(OBJECTS)
	$(CXX) -o $(EXE)  $(OBJECTS) $(CXXFLAGS) nntpd.cpp $(LD_FLAGS)

$(TOOL): $(OBJECTS)
	$(CXX) -o $(TOOL) $(OBJECTS) $(CXXFLAGS) tool.cpp  $(LD_FLAGS)

build-test: $(OBJECTS)
	$(CXX) -o test $(OBJECTS) $(CXXFLAGS) test.cpp $(LD_FLAGS)

test: build-test
	./test

%.o: src/%.cpp
	$(CXX) $(CXXFLAGS) -c -o $@

clean:
	rm -f $(OBJECTS) $(EXE) $(TOOL) test
