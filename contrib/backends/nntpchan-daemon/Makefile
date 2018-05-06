
REPO=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
NNTPCHAN_PATH = $(REPO)/libnntpchan

NNTPCHAN_SRC := $(wildcard $(NNTPCHAN_PATH)/*.cpp)
NNTPCHAN_HDR := $(wildcard $(NNTPCHAN_PATH)/*.hpp)
NNTPCHAN_OBJ := $(NNTPCHAN_SRC:.cpp=.o)

HEADERS_PATH=$(REPO)/include

TOOL_PATH := $(REPO)/tools

TOOL_SRC := $(wildcard $(TOOL_PATH)/*.cpp)
TOOLS := $(TOOL_SRC:.cpp=)

SRCS = $(NNTPCHAN_SRC) $(TOOL_SRC)

OBJ := $(NNTPCHAN_OBJ)

TEST = $(REPO)/test

DAEMON_SRC = $(REPO)/daemon

LD_FLAGS ?= 

INC_FLAGS = -I$(HEADERS_PATH)

ifeq ($(shell uname -s),FreeBSD)
	LD_FLAGS += -lc++fs
	INC_FLAGS += /usr/local/include
	LD_FLAGS += /usr/local/lib/libsodium.a
else 
	LD_FLAGS += -lstdc++fs
	INC_FLAGS += $(shell pkg-config --cflags libsodium)
	LD_FLAGS += $(shell pkg-config --libs --static libsodium)
endif

REQUIRED_CXXFLAGS = -std=c++17 -Wall -Wextra -Werror -pedantic $(INC_FLAGS)

DEBUG = 1

ifeq ($(DEBUG),1)
	REQUIRED_CXXFLAGS += -g
endif

CXXFLAGS += $(REQUIRED_CXXFLAGS)

NNTPCHAN_LIB = $(REPO)/libnntpchan.a

LIBS = $(NNTPCHAN_LIB)

EXE = $(REPO)/nntpd


all: build

format:
	clang-format -i $(SRCS)

build: $(EXE) tools

$(NNTPCHAN_LIB): $(NNTPCHAN_OBJ)
	$(AR) -r $(NNTPCHAN_LIB) $(NNTPCHAN_OBJ)

$(EXE): $(LIBS)
	$(CXX) $(CXXFLAGS) $(DAEMON_SRC)/main.cpp $(LIBS) $(LD_FLAGS) -o $(EXE)

tools: $(TOOLS)

$(TOOLS): $(LIBS)
	$(CXX) $(CXXFLAGS) $@.cpp $(LIBS) $(LD_FLAGS) -o $@

build-test: $(LIBS)
	$(CXX) -o $(TEST) $(CXXFLAGS) test.cpp $(LIBS) $(LD_FLAGS)

test: build-test
	$(TEST)

clean:
	$(RM) $(OBJ) $(LIBS) $(EXE) $(TOOLS) $(TEST)
