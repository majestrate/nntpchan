
REPO=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SRC_PATH = $(REPO)/src

SOURCES := $(wildcard $(SRC_PATH)/*.cpp)
HEADERS := $(wildcard $(SRC_PATH)/*.hpp)
OBJECTS := $(SOURCES:.cpp=.o)

TOOL_SRC_PATH := $(REPO)/tools

TOOL_SRC := $(wildcard $(TOOL_SRC_PATH)/*.cpp)
TOOLS := $(TOOL_SRC:.cpp=)

DAEMON_SRC = $(REPO)/daemon

PKGS := libuv libsodium

LD_FLAGS := $(shell pkg-config --libs $(PKGS))
INC_FLAGS := $(shell pkg-config --cflags $(PKGS)) -I $(REPO)/src
CXXFLAGS := -std=c++11 -Wall -Wextra $(INC_FLAGS)

ifeq ($(DEBUG),1)
	CXXFLAGS += -g
endif

LIB = $(REPO)/libnntpchan.a

EXE = $(REPO)/nntpd


all: $(EXE) $(TOOLS)

$(LIB): $(OBJECTS)
	$(AR) -r $(LIB) $(OBJECTS)

$(EXE): $(LIB)
	$(CXX) $(CXXFLAGS) $(DAEMON_SRC)/main.cpp $(LIB) $(LD_FLAGS) -o $(EXE)

$(TOOL_SRC): $(LIB)

$(TOOLS): $(TOOL_SRC)
	$(CXX)  $(CXXFLAGS) $< $(LIB) $(LD_FLAGS) -o $@

build-test: $(LIB)
	$(CXX) -o test $(CXXFLAGS) test.cpp $(LIB) $(LD_FLAGS)

test: build-test
	./test

%.o: src/%.cpp
	$(CXX) $(CXXFLAGS) -c -o $@

clean:
	rm -f $(OBJECTS) $(LIB) $(EXE) $(TOOLS)
