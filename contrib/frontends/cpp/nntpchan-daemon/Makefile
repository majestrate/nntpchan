
EXE = nntpd

TOOL = nntpchan-tool

CXX = clang++

SRC_PATH = ./src

SOURCES := $(wildcard $(SRC_PATH)/*.cpp)
HEADERS := $(wildcard $(SRC_PATH)/*.hpp)
OBJECTS := $(SOURCES:.cpp=.o)

PKGS := libuv libsodium

LD_FLAGS := $(shell pkg-config --libs $(PKGS))
INC_FLAGS := -I $(SRC_PATH) $(shell pkg-config --cflags $(PKGS))
CXXFLAGS := -std=c++14 -Wall -Wextra $(INC_FLAGS) -g



all: $(EXE) $(TOOL)

$(EXE): $(OBJECTS)	
	$(CXX) -o $(EXE) $(LD_FLAGS) $(OBJECTS) $(CXXFLAGS) nntpchan.cpp

$(TOOL): $(OBJECTS)
	$(CXX) -o $(TOOL) $(LD_FLAGS) $(OBJECTS) $(CXXFLAGS) tool.cpp

%.o: src/%.cpp
	$(CXX) $(CXXFLAGS) -c -o $@ 

clean:
	rm -f $(OBJECTS) $(EXE) $(TOOL)
