
EXE = nntpchan

CXX = clang++

SRC_PATH = ./src

SOURCES := $(wildcard $(SRC_PATH)/*.cpp)
HEADERS := $(wildcard $(SRC_PATH)/*.hpp)
OBJECTS := $(SOURCES:.cpp=.o)

PKGS := libuv

LD_FLAGS := $(shell pkg-config --libs $(PKGS))
INC_FLAGS := -I $(SRC_PATH) $(shell pkg-config --cflags $(PKGS))
CXXFLAGS := -std=c++14 -Wall -Wextra $(INC_FLAGS)



all: $(EXE)

$(EXE): $(OBJECTS)
	$(CXX) -o $(EXE) $(LD_FLAGS) $(OBJECTS)

%.o: src/%.cpp
	$(CXX) $(CXXFLAGS) -c -o $@ 

clean:
	rm -f $(OBJECTS) $(EXE)
