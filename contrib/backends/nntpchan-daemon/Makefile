
REPO=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
NNTPCHAN_PATH = $(REPO)/libnntpchan

NNTPCHAN_SRC := $(wildcard $(NNTPCHAN_PATH)/*.cpp)
NNTPCHAN_HDR := $(wildcard $(NNTPCHAN_PATH)/*.hpp)
NNTPCHAN_OBJ := $(NNTPCHAN_SRC:.cpp=.o)

MUSTACHE_PATH = $(REPO)/libmustache

MUSTACHE_SRC := $(wildcard $(MUSTACHE_PATH)/*.cpp)
MUSTACHE_SRC += $(wildcard $(MUSTACHE_PATH)/*/*.cpp)

MUSTACHE_HDR := $(wildcard $(MUSTACHE_PATH)/*.hpp)
MUSTACHE_OBJ := $(MUSTACHE_SRC:.cpp=.o)

HEADERS_PATH=$(REPO)/include

TOOL_PATH := $(REPO)/tools

TOOL_SRC := $(wildcard $(TOOL_PATH)/*.cpp)
TOOLS := $(TOOL_SRC:.cpp=)

OBJ := $(NNTPCHAN_OBJ) $(MUSTACHE_OBJ)

DAEMON_SRC = $(REPO)/daemon

PKGS := libuv libsodium

LD_FLAGS := $(shell pkg-config --libs $(PKGS)) -lstdc++fs
INC_FLAGS := $(shell pkg-config --cflags $(PKGS)) -I$(HEADERS_PATH)
CXXFLAGS := -std=c++17 -Wall -Wextra -pedantic $(INC_FLAGS)

ifeq ($(DEBUG),1)
	CXXFLAGS += -g
endif

NNTPCHAN_LIB = $(REPO)/libnntpchan.a
MUSTACHE_LIB = $(REPO)/libmustache.a

LIBS = $(NNTPCHAN_LIB) $(MUSTACHE_LIB)

EXE = $(REPO)/nntpd


all: build

build: $(EXE) $(TOOLS)

$(MUSTACHE_LIB): $(MUSTACHE_OBJ)
	$(AR) -r $(MUSTACHE_LIB) $(MUSTACHE_OBJ)

$(NNTPCHAN_LIB): $(NNTPCHAN_OBJ)
	$(AR) -r $(NNTPCHAN_LIB) $(NNTPCHAN_OBJ)

$(EXE): $(LIBS)
	$(CXX) $(CXXFLAGS) $(DAEMON_SRC)/main.cpp $(NNTPCHAN_LIB) $(LD_FLAGS) -o $(EXE)

$(TOOL_SRC): $(NNTPCHAN_LIB)

$(TOOLS): $(TOOL_SRC)
	$(CXX)  $(CXXFLAGS) $< $(NNTPCHAN_LIB) $(LD_FLAGS) -o $@

build-test: $(LIB)
	$(CXX) -o $(REPO)/test $(CXXFLAGS) test.cpp $(NNTPCHAN_LIB) $(LD_FLAGS)

test: build-test
	$(REPO)/test

clean:
	rm -f $(OBJ) $(NNTPCHAN_LIB) $(EXE) $(TOOLS)
