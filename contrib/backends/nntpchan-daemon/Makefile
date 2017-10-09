
REPO=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SRC_PATH = $(REPO)/libnntpchan

SOURCES := $(wildcard $(SRC_PATH)/*.cpp)
HEADERS := $(wildcard $(SRC_PATH)/*.hpp)
OBJECTS := $(SOURCES:.cpp=.o)

TOOL_SRC_PATH := $(REPO)/tools

TOOL_SRC := $(wildcard $(TOOL_SRC_PATH)/*.cpp)
TOOLS := $(TOOL_SRC:.cpp=)

DAEMON_SRC = $(REPO)/daemon

PKGS := libuv libsodium

LD_FLAGS := $(shell pkg-config --libs $(PKGS)) -lstdc++fs
INC_FLAGS := $(shell pkg-config --cflags $(PKGS)) -I$(SRC_PATH)
CXXFLAGS := -std=c++17 -Wall -Wextra -pedantic $(INC_FLAGS)

ifeq ($(DEBUG),1)
	CXXFLAGS += -g
endif

LIB = $(REPO)/libnntpchan.a

EXE = $(REPO)/nntpd


all: build

build: $(EXE) $(TOOLS)

$(LIB): $(OBJECTS)
	$(AR) -r $(LIB) $(OBJECTS)

$(EXE): $(LIB)
	$(CXX) $(CXXFLAGS) $(DAEMON_SRC)/main.cpp $(LIB) $(LD_FLAGS) -o $(EXE)

$(TOOL_SRC): $(LIB)

$(TOOLS): $(TOOL_SRC)
	$(CXX)  $(CXXFLAGS) $< $(LIB) $(LD_FLAGS) -o $@

build-test: $(LIB)
	$(CXX) -o $(REPO)/test $(CXXFLAGS) test.cpp $(LIB) $(LD_FLAGS)

test: build-test
	$(REPO)/test

%.o: src/%.cpp
	$(CXX) $(CXXFLAGS) -c -o $@

clean:
	rm -f $(OBJECTS) $(LIB) $(EXE) $(TOOLS)
