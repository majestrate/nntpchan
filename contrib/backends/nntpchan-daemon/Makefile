
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

ifeq ($(shell uname -s) "FreeBSD")
	SODIUM_INC = /usr/local/include
	SODIUM_LIB = /usr/local/lib/libsodium.a
else 
	SODIUM_INC = $(shell pkg-config --cflags libsodium)
	SODIUM_LIB = $(shell pkg-config --libs --static libsodium)	
endif

LD_FLAGS :=  -lstdc++fs $(SODIUM_LIB)
INC_FLAGS := -I$(HEADERS_PATH) $(SODIUM_INC)
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
