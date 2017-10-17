
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

OBJ := $(NNTPCHAN_OBJ)
OBJ += $(MUSTACHE_OBJ)

TEST = $(REPO)/test

DAEMON_SRC = $(REPO)/daemon

PKGS := libuv libsodium

LD_FLAGS := $(shell pkg-config --libs $(PKGS)) -lstdc++fs
INC_FLAGS := $(shell pkg-config --cflags $(PKGS)) -I$(HEADERS_PATH)
REQUIRED_CXXFLAGS = -std=c++17 -Wall -Wextra -Werror -pedantic $(INC_FLAGS)

DEBUG = 1

ifeq ($(DEBUG),1)
	REQUIRED_CXXFLAGS += -g
endif

CXXFLAGS += $(REQUIRED_CXXFLAGS)

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
	$(CXX) $(CXXFLAGS) $(DAEMON_SRC)/main.cpp $(LIBS) $(LD_FLAGS) -o $(EXE)

$(TOOLS): $(TOOL_SRC) $(LIBS)
	$(CXX) $(CXXFLAGS) $< $(LIBS) $(LD_FLAGS) -o $@

build-test: $(LIBS)
	$(CXX) -o $(TEST) $(CXXFLAGS) test.cpp $(LIBS) $(LD_FLAGS)

test: build-test
	$(TEST)

clean:
	$(RM) $(OBJ) $(LIBS) $(EXE) $(TOOLS) $(TEST)
